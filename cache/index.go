package cache

// NewUserIndex returns a new instance of a user index
func NewUserIndex() *UserIndex {
	return &UserIndex{}
}

// AddUser inserts a new user into the index. The *Data struct only contains the nickname.)
func (index UserIndex) AddUser(nick string, url string) {
	imutex.Lock()
	index[url] = &Data{nick: nick}
	imutex.Unlock()
}

// DelUser removes a user from the index completely.
func (index UserIndex) DelUser(url string) {
	imutex.Lock()
	delete(index, url)
	imutex.Unlock()
}
