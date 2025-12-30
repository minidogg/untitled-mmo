package main

const (
	TicksPerSecond = 30

	Gravity      = 12.0
	Accel        = 8.0
	Friction     = 7.0
	MaxSpeedX    = 30.0
	JumpImpulse  = 90.0
	MaxFallSpeed = 120.0
)

type Input struct {
	Left  bool
	Right bool
	Jump  bool
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

func Step(p *Entity, input Input, tiles TileMap) {
	if input.Left {
		p.Velocity.X -= Accel
	}
	if input.Right {
		p.Velocity.X += Accel
	}

	p.Velocity.X = Clamp(p.Velocity.X, -MaxSpeedX, MaxSpeedX)

	if !input.Left && !input.Right {
		p.Velocity.X = (p.Velocity.X * Friction) / 100
	}

	p.Velocity.Y += Gravity
	if p.Velocity.Y > MaxFallSpeed {
		p.Velocity.Y = MaxFallSpeed
	}

	if input.Jump && p.Grounded {
		p.Velocity.Y = -JumpImpulse
		p.Grounded = false
	}

	p.Position.X += p.Velocity.X
	p.Position.Y += p.Velocity.Y
}