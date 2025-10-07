package custom_resolver

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	lokstraConfig "github.com/primadi/lokstra/core/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sSecretResolver reads secrets from Kubernetes
// Supports both in-cluster config and kubeconfig file
//
// Key format: "namespace/secret-name/key-name"
// Example: "default/database-credentials/password"
//
// If namespace is omitted, uses default namespace or POD_NAMESPACE env var
// Example: "database-credentials/password"
type K8sSecretResolver struct {
	client    *kubernetes.Clientset
	cache     map[string]string
	namespace string // default namespace
}

// NewK8sSecretResolver creates a Kubernetes secret resolver
// Attempts in-cluster config first, falls back to kubeconfig
func NewK8sSecretResolver() (*K8sSecretResolver, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		config, err = loadKubeconfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Get default namespace from POD_NAMESPACE env or use "default"
	defaultNamespace := os.Getenv("POD_NAMESPACE")
	if defaultNamespace == "" {
		defaultNamespace = "default"
	}

	return &K8sSecretResolver{
		client:    clientset,
		cache:     make(map[string]string),
		namespace: defaultNamespace,
	}, nil
}

// NewK8sSecretResolverWithNamespace creates resolver with custom default namespace
func NewK8sSecretResolverWithNamespace(namespace string) (*K8sSecretResolver, error) {
	resolver, err := NewK8sSecretResolver()
	if err != nil {
		return nil, err
	}
	resolver.namespace = namespace
	return resolver, nil
}

// Resolve retrieves secret value from Kubernetes
// Key format: "namespace/secret-name/key-name" or "secret-name/key-name"
func (r *K8sSecretResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
	// Check cache first
	cacheKey := r.makeCacheKey(key)
	if value, ok := r.cache[cacheKey]; ok {
		return value, true
	}

	// Parse key: namespace/secret-name/key-name or secret-name/key-name
	namespace, secretName, keyName, err := r.parseKey(key)
	if err != nil {
		log.Printf("Invalid K8s secret key format %q: %v (using default)", key, err)
		return defaultValue, false
	}

	// Get secret from Kubernetes
	secret, err := r.client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Printf("Failed to get K8s secret %s/%s: %v (using default)", namespace, secretName, err)
		return defaultValue, false
	}

	// Get the specific key from secret data
	value, ok := secret.Data[keyName]
	if !ok {
		log.Printf("Key %q not found in K8s secret %s/%s (using default)", keyName, namespace, secretName)
		return defaultValue, false
	}

	secretValue := string(value)

	// Cache the value
	r.cache[cacheKey] = secretValue
	return secretValue, true
}

// ResolveEntireSecret retrieves entire secret as map
func (r *K8sSecretResolver) ResolveEntireSecret(namespace, secretName string) (map[string]string, error) {
	secret, err := r.client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	result := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		result[k] = string(v)
	}
	return result, nil
}

// ListSecrets lists all secrets in namespace
func (r *K8sSecretResolver) ListSecrets(namespace string) ([]string, error) {
	if namespace == "" {
		namespace = r.namespace
	}

	secrets, err := r.client.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	names := make([]string, len(secrets.Items))
	for i, secret := range secrets.Items {
		names[i] = secret.Name
	}
	return names, nil
}

// ClearCache clears the internal cache
func (r *K8sSecretResolver) ClearCache() {
	r.cache = make(map[string]string)
}

// parseKey parses key format: "namespace/secret-name/key-name" or "secret-name/key-name"
func (r *K8sSecretResolver) parseKey(key string) (namespace, secretName, keyName string, err error) {
	parts := strings.Split(key, "/")

	switch len(parts) {
	case 2:
		// Format: secret-name/key-name
		return r.namespace, parts[0], parts[1], nil
	case 3:
		// Format: namespace/secret-name/key-name
		return parts[0], parts[1], parts[2], nil
	default:
		return "", "", "", fmt.Errorf("invalid format, expected 'secret-name/key' or 'namespace/secret-name/key', got %q", key)
	}
}

// makeCacheKey creates a unique cache key
func (r *K8sSecretResolver) makeCacheKey(key string) string {
	namespace, secretName, keyName, err := r.parseKey(key)
	if err != nil {
		return key
	}
	return fmt.Sprintf("%s/%s/%s", namespace, secretName, keyName)
}

// loadKubeconfig loads kubernetes config from kubeconfig file
func loadKubeconfig() (*rest.Config, error) {
	// Try KUBECONFIG env var first
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		// Try default location
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfigPath, err)
	}

	return config, nil
}

// WatchSecret sets up a watch on a secret (for advanced use cases)
// Returns a channel that receives updates when secret changes
func (r *K8sSecretResolver) WatchSecret(namespace, secretName string) (<-chan *corev1.Secret, error) {
	if namespace == "" {
		namespace = r.namespace
	}

	watcher, err := r.client.CoreV1().Secrets(namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", secretName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to watch secret: %w", err)
	}

	ch := make(chan *corev1.Secret)
	go func() {
		defer close(ch)
		defer watcher.Stop()

		for event := range watcher.ResultChan() {
			if secret, ok := event.Object.(*corev1.Secret); ok {
				ch <- secret
			}
		}
	}()

	return ch, nil
}

var _ lokstraConfig.VariableResolver = (*K8sSecretResolver)(nil)

// Example usage
func ExampleK8sSecretResolver() {
	// Register Kubernetes Secret resolver
	k8sResolver, err := NewK8sSecretResolver()
	if err != nil {
		log.Fatal(err)
	}
	lokstraConfig.AddVariableResolver("K8S", k8sResolver)

	// Now you can use ${@K8S:...} in your YAML configs
	//
	// Example 1: Using default namespace
	// config/production.yaml
	//
	// services:
	//   - name: postgres
	//     type: dbpool_pg
	//     config:
	//       # Reads from secret "database-credentials" in default namespace
	//       host: ${@K8S:database-credentials/host}
	//       port: ${@K8S:database-credentials/port}
	//       user: ${@K8S:database-credentials/username}
	//       password: ${@K8S:database-credentials/password}
	//       database: ${@K8S:database-credentials/database}
	//
	//   - name: api
	//     config:
	//       # Reads from secret "api-secrets" in default namespace
	//       apiKey: ${@K8S:api-secrets/api-key}
	//       jwtSecret: ${@K8S:api-secrets/jwt-secret}
	//
	// Example 2: Specify namespace explicitly
	//
	// services:
	//   - name: redis
	//     config:
	//       # Reads from secret "redis-auth" in "production" namespace
	//       url: ${@K8S:production/redis-auth/url}
	//       password: ${@K8S:production/redis-auth/password}
	//
	// Example 3: With fallback defaults
	//
	// servers:
	//   - name: api
	//     # Falls back to http://localhost:8080 if secret not found
	//     baseUrl: ${@K8S:app-config/base-url:http://localhost:8080}
}

// Example: Create Kubernetes secret from command line
//
// kubectl create secret generic database-credentials \
//   --from-literal=host=postgres.default.svc.cluster.local \
//   --from-literal=port=5432 \
//   --from-literal=username=myuser \
//   --from-literal=password=mypassword \
//   --from-literal=database=mydb
//
// kubectl create secret generic api-secrets \
//   --from-literal=api-key=sk_live_abc123 \
//   --from-literal=jwt-secret=super-secret-jwt-key
//
// Or from file:
//
// kubectl create secret generic app-config \
//   --from-file=config.json=./config.json \
//   --from-file=tls.crt=./tls.crt \
//   --from-file=tls.key=./tls.key
