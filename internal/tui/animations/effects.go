package animations

import (
	"math"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Animation provides smooth transitions and effects
type Animation struct {
	StartTime    time.Time
	Duration     time.Duration
	EasingFunc   EasingFunction
	FromValue    float64
	ToValue      float64
	CurrentValue float64
	Completed    bool
}

// EasingFunction defines animation curves
type EasingFunction func(t float64) float64

// Common easing functions
var (
	EaseLinear     = func(t float64) float64 { return t }
	EaseInQuad     = func(t float64) float64 { return t * t }
	EaseOutQuad    = func(t float64) float64 { return t * (2 - t) }
	EaseInOutQuad  = func(t float64) float64 { return t * t * (3 - 2*t) }
	EaseInCubic    = func(t float64) float64 { return t * t * t }
	EaseOutCubic   = func(t float64) float64 { return 1 - math.Pow(1-t, 3) }
	EaseInOutCubic = func(t float64) float64 {
		if t < 0.5 {
			return 4 * t * t * t
		}
		return 1 - math.Pow(-2*t+2, 3)/2
	}
	EaseBounce = func(t float64) float64 {
		if t < 1/2.75 {
			return 7.5625 * t * t
		} else if t < 2/2.75 {
			t -= 1.5 / 2.75
			return 7.5625*t*t + 0.75
		} else if t < 2.5/2.75 {
			t -= 2.25 / 2.75
			return 7.5625*t*t + 0.9375
		} else {
			t -= 2.625 / 2.75
			return 7.5625*t*t + 0.984375
		}
	}
)

// NewAnimation creates a new animation
func NewAnimation(duration time.Duration, from, to float64, easing EasingFunction) *Animation {
	if easing == nil {
		easing = EaseOutQuad
	}

	return &Animation{
		StartTime:  time.Now(),
		Duration:   duration,
		EasingFunc: easing,
		FromValue:  from,
		ToValue:    to,
		Completed:  false,
	}
}

// Update advances the animation
func (a *Animation) Update() {
	if a.Completed {
		return
	}

	elapsed := time.Since(a.StartTime)
	progress := float64(elapsed) / float64(a.Duration)

	if progress >= 1.0 {
		a.CurrentValue = a.ToValue
		a.Completed = true
		return
	}

	// Apply easing function
	easedProgress := a.EasingFunc(progress)
	a.CurrentValue = a.FromValue + (a.ToValue-a.FromValue)*easedProgress
}

// IsCompleted checks if animation is finished
func (a *Animation) IsCompleted() bool {
	return a.Completed
}

// GetValue returns the current animated value
func (a *Animation) GetValue() float64 {
	return a.CurrentValue
}

// Reset restarts the animation
func (a *Animation) Reset() {
	a.StartTime = time.Now()
	a.Completed = false
	a.CurrentValue = a.FromValue
}

// PulseAnimation creates a pulsing effect
type PulseAnimation struct {
	*Animation
	frequency float64
}

// NewPulseAnimation creates a pulsing animation
func NewPulseAnimation(frequency float64) *PulseAnimation {
	return &PulseAnimation{
		Animation: NewAnimation(time.Second, 0, 1, EaseInOutQuad),
		frequency: frequency,
	}
}

// UpdatePulse updates the pulse animation
func (p *PulseAnimation) UpdatePulse() {
	p.Update()
	if p.Completed {
		// Reverse direction for continuous pulse
		p.FromValue, p.ToValue = p.ToValue, p.FromValue
		p.Reset()
	}
}

// ColorTransition creates smooth color transitions
type ColorTransition struct {
	FromColor lipgloss.Color
	ToColor   lipgloss.Color
	*Animation
}

// NewColorTransition creates a color transition animation
func NewColorTransition(from, to lipgloss.Color, duration time.Duration) *ColorTransition {
	return &ColorTransition{
		FromColor: from,
		ToColor:   to,
		Animation: NewAnimation(duration, 0, 1, EaseInOutQuad),
	}
}

// GetCurrentColor returns the interpolated color
func (ct *ColorTransition) GetCurrentColor() lipgloss.Color {
	if ct.Completed {
		return ct.ToColor
	}

	progress := ct.GetValue()
	// Simple interpolation - could be enhanced with proper color interpolation
	if progress > 0.5 {
		return ct.ToColor
	}
	return ct.FromColor
}

// FadeAnimation creates fade in/out effects
type FadeAnimation struct {
	*Animation
	style lipgloss.Style
}

// NewFadeAnimation creates a fade animation
func NewFadeAnimation(duration time.Duration, fadeIn bool) *FadeAnimation {
	from := 1.0
	to := 0.0
	if fadeIn {
		from, to = 0.0, 1.0
	}

	return &FadeAnimation{
		Animation: NewAnimation(duration, from, to, EaseInOutQuad),
	}
}

// ApplyToStyle applies the fade effect to a style
func (fa *FadeAnimation) ApplyToStyle(baseStyle lipgloss.Style) lipgloss.Style {
	opacity := fa.GetValue()

	// Apply opacity to foreground color
	// This is a simplified version - real implementation would need color manipulation
	if opacity < 0.5 {
		return baseStyle.Foreground(lipgloss.Color("#666666"))
	}
	return baseStyle
}

// SlideAnimation creates sliding effects
type SlideAnimation struct {
	*Animation
	direction string // "up", "down", "left", "right"
	distance  int
}

// NewSlideAnimation creates a slide animation
func NewSlideAnimation(direction string, distance int, duration time.Duration) *SlideAnimation {
	return &SlideAnimation{
		Animation: NewAnimation(duration, 0, float64(distance), EaseOutQuad),
		direction: direction,
		distance:  distance,
	}
}

// GetOffset returns the current slide offset
func (sa *SlideAnimation) GetOffset() (x, y int) {
	offset := int(sa.GetValue())

	switch sa.direction {
	case "up":
		return 0, -offset
	case "down":
		return 0, offset
	case "left":
		return -offset, 0
	case "right":
		return offset, 0
	default:
		return 0, 0
	}
}

// ScaleAnimation creates zoom effects
type ScaleAnimation struct {
	*Animation
}

// NewScaleAnimation creates a scale animation
func NewScaleAnimation(fromScale, toScale float64, duration time.Duration) *ScaleAnimation {
	return &ScaleAnimation{
		Animation: NewAnimation(duration, fromScale, toScale, EaseOutCubic),
	}
}

// GetScale returns the current scale factor
func (sa *ScaleAnimation) GetScale() float64 {
	return sa.GetValue()
}

// StaggerAnimation creates staggered animations for multiple elements
type StaggerAnimation struct {
	animations []*Animation
	delay      time.Duration
}

// NewStaggerAnimation creates a staggered animation
func NewStaggerAnimation(count int, duration, delay time.Duration, from, to float64, easing EasingFunction) *StaggerAnimation {
	sa := &StaggerAnimation{
		animations: make([]*Animation, count),
		delay:      delay,
	}

	for i := 0; i < count; i++ {
		sa.animations[i] = NewAnimation(duration, from, to, easing)
	}

	return sa
}

// UpdateStagger updates all staggered animations
func (sa *StaggerAnimation) UpdateStagger() {
	now := time.Now()
	for i, anim := range sa.animations {
		delay := time.Duration(i) * sa.delay
		if now.After(anim.StartTime.Add(delay)) {
			anim.Update()
		}
	}
}

// GetValue gets the value for a specific animation
func (sa *StaggerAnimation) GetValue(index int) float64 {
	if index >= 0 && index < len(sa.animations) {
		return sa.animations[index].GetValue()
	}
	return 0
}

// AnimationManager manages multiple animations
type AnimationManager struct {
	animations map[string]interface{}
}

// NewAnimationManager creates a new animation manager
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		animations: make(map[string]interface{}),
	}
}

// AddAnimation adds an animation with a key
func (am *AnimationManager) AddAnimation(key string, anim interface{}) {
	am.animations[key] = anim
}

// GetAnimation gets an animation by key
func (am *AnimationManager) GetAnimation(key string) interface{} {
	return am.animations[key]
}

// UpdateAll updates all animations
func (am *AnimationManager) UpdateAll() {
	for _, anim := range am.animations {
		switch a := anim.(type) {
		case *Animation:
			a.Update()
		case *PulseAnimation:
			a.UpdatePulse()
		case *ColorTransition:
			a.Update()
		case *FadeAnimation:
			a.Update()
		case *SlideAnimation:
			a.Update()
		case *ScaleAnimation:
			a.Update()
		case *StaggerAnimation:
			a.UpdateStagger()
		}
	}
}

// CleanCompleted removes completed animations
func (am *AnimationManager) CleanCompleted() {
	for key, anim := range am.animations {
		switch a := anim.(type) {
		case *Animation:
			if a.Completed {
				delete(am.animations, key)
			}
		}
	}
}
