package main

const (
	TicksPerSecond = 30

	Gravity      = -0.3
	Accel        = 8.0
	Friction     = 7.0
	MaxSpeedX    = 30.0
	JumpImpulse  = 90.0
	MaxFallSpeed = 10.0
)

type Input struct {
	Left  bool `json:"left"`
	Right bool `json:"right"`
	Jump  bool `json:"jump"`
}

func Clamp(val, min, max float32) float32 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func (p *Entity) StepPhysics() {
	if p.Input.Left {
		p.Velocity.X -= Accel
	}
	if p.Input.Right {
		p.Velocity.X += Accel
	}

	p.Velocity.X = Clamp(p.Velocity.X, -MaxSpeedX, MaxSpeedX)

	if !p.Input.Left && !p.Input.Right {
		p.Velocity.X = (p.Velocity.X * Friction) / 100
	}

	p.Velocity.Y += Gravity
	if p.Velocity.Y > MaxFallSpeed {
		p.Velocity.Y = MaxFallSpeed
	}

	if p.Input.Jump && p.State.Grounded {
		p.Velocity.Y = -JumpImpulse
		p.State.Grounded = false
	}

	p.Position.X += p.Velocity.X
	p.Position.Y += p.Velocity.Y
}
