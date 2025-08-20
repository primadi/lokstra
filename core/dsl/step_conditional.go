package dsl

import "fmt"

// stepConditional executes a step only if condition is met
type stepConditional[TParam any] struct {
	condition func(*FlowContext[TParam]) bool
	thenStep  Step[TParam]
	elseStep  Step[TParam]
	BaseStep[TParam]
}

func newStepConditional[TParam any](condition func(*FlowContext[TParam]) bool, thenStep Step[TParam]) *stepConditional[TParam] {
	return &stepConditional[TParam]{
		condition: condition,
		thenStep:  thenStep,
		BaseStep: BaseStep[TParam]{
			meta: StepMeta{
				Name: "Conditional Step",
				Kind: StepNormal,
			},
		},
	}
}

func (s *stepConditional[TParam]) Run(ctx *FlowContext[TParam]) error {
	conditionResult := s.condition(ctx)

	// Metrics: Conditional evaluation
	if ctx.serviceVar.Metrics != nil {
		ctx.serviceVar.Metrics.IncCounter("dsl_conditional_evaluated", map[string]string{
			"result": fmt.Sprintf("%t", conditionResult),
		})
	}

	if conditionResult {
		if s.thenStep != nil {
			return s.thenStep.Run(ctx)
		}
	} else {
		if s.elseStep != nil {
			return s.elseStep.Run(ctx)
		}
	}
	return nil
}

func (s *stepConditional[TParam]) Else(elseStep Step[TParam]) Step[TParam] {
	s.elseStep = elseStep
	return s
}

func (s *stepConditional[TParam]) SaveAs(name string) Step[TParam] {
	s.saveAs = name
	return s
}

// stepCustom allows custom function execution
type stepCustom[TParam any] struct {
	fn func(*FlowContext[TParam]) error
	BaseStep[TParam]
}

func newStepCustom[TParam any](fn func(*FlowContext[TParam]) error) *stepCustom[TParam] {
	return &stepCustom[TParam]{
		fn: fn,
		BaseStep: BaseStep[TParam]{
			meta: StepMeta{
				Name: "Custom Function",
				Kind: StepNormal,
			},
		},
	}
}

func (s *stepCustom[TParam]) Run(ctx *FlowContext[TParam]) error {
	return s.fn(ctx)
}

func (s *stepCustom[TParam]) SaveAs(name string) Step[TParam] {
	s.saveAs = name
	return s
}

// stepRetry allows retrying a step with exponential backoff
type stepRetry[TParam any] struct {
	step       Step[TParam]
	maxRetries int
	BaseStep[TParam]
}

func newStepRetry[TParam any](step Step[TParam], maxRetries int) *stepRetry[TParam] {
	return &stepRetry[TParam]{
		step:       step,
		maxRetries: maxRetries,
		BaseStep: BaseStep[TParam]{
			meta: StepMeta{
				Name: "Retry Step",
				Kind: StepNormal,
			},
		},
	}
}

func (s *stepRetry[TParam]) Run(ctx *FlowContext[TParam]) error {
	var lastErr error
	for i := 0; i <= s.maxRetries; i++ {
		// Metrics: Retry attempt
		if ctx.serviceVar.Metrics != nil {
			ctx.serviceVar.Metrics.IncCounter("dsl_retry_attempt", map[string]string{
				"attempt":     fmt.Sprintf("%d", i+1),
				"max_retries": fmt.Sprintf("%d", s.maxRetries+1),
			})
		}

		err := s.step.Run(ctx)
		if err == nil {
			// Metrics: Retry success
			if ctx.serviceVar.Metrics != nil {
				ctx.serviceVar.Metrics.IncCounter("dsl_retry_succeeded", map[string]string{
					"attempts_needed": fmt.Sprintf("%d", i+1),
				})
			}
			return nil
		}
		lastErr = err

		if ctx.serviceVar.Logger != nil {
			ctx.serviceVar.Logger.Warnf("Step retry %d/%d failed: %v", i+1, s.maxRetries+1, err)
		}
	}

	// Metrics: Retry exhausted
	if ctx.serviceVar.Metrics != nil {
		ctx.serviceVar.Metrics.IncCounter("dsl_retry_exhausted", map[string]string{
			"max_retries": fmt.Sprintf("%d", s.maxRetries+1),
		})
	}

	return lastErr
}

func (s *stepRetry[TParam]) SaveAs(name string) Step[TParam] {
	s.saveAs = name
	return s
}
