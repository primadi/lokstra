package request

func (c *Context) QueryParam(name string, defaultValue string) string {
	v := c.R.URL.Query().Get(name)
	if v == "" {
		return defaultValue
	}
	return v
}

func (c *Context) FormParam(name string, defaultValue string) string {
	v := c.R.FormValue(name)
	if v == "" {
		return defaultValue
	}
	return v
}

func (c *Context) PathParam(name string, defaultValue string) string {
	v := c.R.PathValue(name)
	if v == "" {
		return defaultValue
	}
	return v
}

func (c *Context) HeaderParam(name string, defaultValue string) string {
	v := c.R.Header.Get(name)
	if v == "" {
		return defaultValue
	}
	return v
}
