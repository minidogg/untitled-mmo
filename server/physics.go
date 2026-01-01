package main

const (
	TicksPerSecond = 60
	TileSize       = 16

	Gravity      = -0.15
	Accel        = 0.8
	Friction     = 0.82
	MaxSpeedX    = 10.0
	JumpImpulse  = 90.0
	MaxFallSpeed = 5.0
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

func (p *Entity) StepPhysics(cmap CollisionMap) {
	if p.Input.Left {
		p.Velocity.X -= Accel
	}
	if p.Input.Right {
		p.Velocity.X += Accel
	}

	p.Velocity.X = Clamp(p.Velocity.X, -MaxSpeedX, MaxSpeedX)

	if !p.Input.Left && !p.Input.Right {
		p.Velocity.X = p.Velocity.X * Friction
	}

	p.Velocity.Y += Gravity
	if p.Velocity.Y < -MaxFallSpeed {
		p.Velocity.Y = MaxFallSpeed
	}

	if p.Input.Jump && p.State.Grounded {
		p.Velocity.Y = JumpImpulse
		p.State.Grounded = false
	}

	p.Position.X += p.Velocity.X
	p.Position.Y += p.Velocity.Y
	p.State.Grounded = false

	// dont ask me how this works
	tileW, tileH := float32(TileSize), float32(TileSize)
	for _, t := range cmap {
		if CheckCollision(p.Position, p.Size.X, p.Size.Y, t.X, t.Y, tileW, tileH) {
			if p.Velocity.Y > 0 && p.Position.Y+p.Size.Y > t.Y && p.Position.Y < t.Y {
				p.Position.Y = t.Y - p.Size.Y
				p.Velocity.Y = 0
				p.State.Grounded = true
			} else if p.Velocity.Y < 0 && p.Position.Y < t.Y+tileH && p.Position.Y+p.Size.Y > t.Y+tileH {
				p.Position.Y = t.Y + tileH
				p.Velocity.Y = 0
			}

			if p.Velocity.X > 0 && p.Position.X+p.Size.X > t.X && p.Position.X < t.X {
				p.Position.X = t.X - p.Size.X
				p.Velocity.X = 0
			} else if p.Velocity.X < 0 && p.Position.X < t.X+tileW && p.Position.X+p.Size.X > t.X+tileW {
				p.Position.X = t.X + tileW
				p.Velocity.X = 0
			}
		}
	}
}

func CheckCollision(ePos Vec2, eW, eH float32, tX, tY, tW, tH float32) bool {
	return ePos.X < tX+tW &&
		ePos.X+eW > tX &&
		ePos.Y < tY+tH &&
		ePos.Y+eH > tY
}

func BuildCollisionMap(tileLayer TileMapLayer) CollisionMap {
	var cmap CollisionMap

	for _, arr := range tileLayer {
		for i := 0; i+1 < len(arr); i += 2 {
			cmap = append(cmap, Vec2{
				X: arr[i] * TileSize,
				Y: arr[i+1] * TileSize,
			})
		}
	}

	return cmap
}
