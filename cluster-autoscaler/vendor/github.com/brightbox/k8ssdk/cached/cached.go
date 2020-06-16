package cached

import (
	"net/http"
	"time"

	brightbox "github.com/brightbox/gobrightbox"
	cache "github.com/patrickmn/go-cache"
	klog "k8s.io/klog/v2"
)

const (
	expirationTime = 5 * time.Second
	purgeTime      = 30 * time.Second
)

// Client is a cached brightbox Client
type Client struct {
	clientCache *cache.Cache
	brightbox.Client
}

// NewClient creates and returns a cached Client
func NewClient(url string, account string, httpClient *http.Client) (*Client, error) {
	cl, err := brightbox.NewClient(url, account, httpClient)
	if err != nil {
		return nil, err
	}
	return &Client{
		clientCache: cache.New(expirationTime, purgeTime),
		Client:      *cl,
	}, err
}

//Server fetches a server by id
func (c *Client) Server(identifier string) (*brightbox.Server, error) {
	if cachedServer, found := c.clientCache.Get(identifier); found {
		klog.V(4).Infof("Cache hit %q", identifier)
		return cachedServer.(*brightbox.Server), nil
	}
	server, err := c.Client.Server(identifier)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Cacheing %q", identifier)
	c.clientCache.Set(identifier, server, cache.DefaultExpiration)
	return server, nil
}

//ServerGroup fetches a server group by id
func (c *Client) ServerGroup(identifier string) (*brightbox.ServerGroup, error) {
	if cachedServerGroup, found := c.clientCache.Get(identifier); found {
		klog.V(4).Infof("Cache hit %q", identifier)
		return cachedServerGroup.(*brightbox.ServerGroup), nil
	}
	serverGroup, err := c.Client.ServerGroup(identifier)
	if err != nil {
		return nil, err
	}
	klog.V(4).Infof("Cacheing %q", identifier)
	c.clientCache.Set(identifier, serverGroup, cache.DefaultExpiration)
	return serverGroup, nil
}

//DestroyServer removes a server by id
func (c *Client) DestroyServer(identifier string) error {
	err := c.Client.DestroyServer(identifier)
	if err == nil {
		c.clientCache.Delete(identifier)
	}
	return err
}

//DestroyServerGroup removes a server group by id
func (c *Client) DestroyServerGroup(identifier string) error {
	err := c.Client.DestroyServerGroup(identifier)
	if err == nil {
		c.clientCache.Delete(identifier)
	}
	return err
}
