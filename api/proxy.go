package api

type (
	Proxy interface {
		Create(c *Create) (any, error)
		Read(r *Read) (any, error)
		Update(u *Update) (any, error)
		Delete(d *Delete) error
		Search(c *Search) (any, error)
		Patch(p *Patch) (any, error)
	}
)
