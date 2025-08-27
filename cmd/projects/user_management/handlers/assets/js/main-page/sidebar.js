// Sidebar Management - User Management Application
// Handles sidebar active menu updates and dropdown states

// Update sidebar active menu based on current path
function updateSidebarActiveMenu(targetPath) {
  // Use provided path or fallback to current location
  const currentPath = targetPath || window.location.pathname;
  console.log("📍 Updating sidebar for path:", currentPath);
  console.log("🎯 Target path provided:", targetPath || "none");
  console.log("🌐 Window location:", window.location.pathname);

  // Cache DOM queries for better performance
  const sidebarLinks = document.querySelectorAll(".sidebar-nav-link");
  console.log("🔍 Found", sidebarLinks.length, "sidebar links");

  // Debug: Log all available links
  sidebarLinks.forEach((link, index) => {
    const href = link.getAttribute("href");
    console.log("  Link " + index + ': href="' + href + '"');
  });

  // Batch DOM operations to minimize reflows
  const updateOperations = [];

  // Remove active styles from all links first
  sidebarLinks.forEach((link) => {
    updateOperations.push(() => {
      link.classList.remove(
        "bg-gray-700",
        "border",
        "border-gray-600",
        "text-white"
      );
      link.classList.add("text-gray-300");
    });
  });

  // Find best matching link
  let bestMatch = null;
  let bestMatchLength = 0;

  sidebarLinks.forEach((link) => {
    const href = link.getAttribute("href");
    if (href) {
      // Check for exact match first
      if (currentPath === href) {
        console.log('✅ Exact match found: href="' + href + '"');
        bestMatch = { link, exact: true, length: href.length };
        bestMatchLength = href.length;
      }
      // Check for partial match (longer matches have priority)
      else if (href !== "/" && currentPath.startsWith(href)) {
        if (href.length > bestMatchLength) {
          console.log(
            '🎯 Partial match found: href="' +
              href +
              '" (length=' +
              href.length +
              ")"
          );
          bestMatch = { link, exact: false, length: href.length };
          bestMatchLength = href.length;
        } else {
          console.log(
            '⚠️ Partial match rejected: href="' +
              href +
              '" (length=' +
              href.length +
              " <= " +
              bestMatchLength +
              ")"
          );
        }
      } else {
        console.log('❌ No match: href="' + href + '"');
      }
    }
  });

  if (bestMatch) {
    console.log(
      '🎉 Best match selected: href="' +
        bestMatch.link.getAttribute("href") +
        '" (exact=' +
        bestMatch.exact +
        ")"
    );
  } else {
    console.log("😞 No match found for path: " + currentPath);
  }

  // Add active styles to matching link
  if (bestMatch) {
    const { link } = bestMatch;
    updateOperations.push(() => {
      link.classList.add(
        "bg-gray-700",
        "border",
        "border-gray-600",
        "text-white"
      );
      link.classList.remove("text-gray-300");

      // Handle dropdown if needed
      const dropdown = link.closest("li[x-data]");
      if (dropdown && dropdown._x_dataStack) {
        dropdown._x_dataStack[0].menuOpen = true;
      }
    });
  }

  // Execute all DOM updates in a single animation frame
  requestAnimationFrame(() => {
    updateOperations.forEach((operation) => operation());
    console.log(
      "✅ Sidebar updated efficiently with",
      updateOperations.length,
      "operations"
    );
  });
}
