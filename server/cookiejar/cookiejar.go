package cookiejar

import (
	"net/http"
	"net/url"
)

// CookieJar ...
type CookieJar struct {
	Jar map[string][]*http.Cookie
}

// SetCookies ...
func (p *CookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	// fmt.Printf("The URL is : %s\n", u.String())
	// fmt.Printf("The cookie being set is : %s\n", cookies)
	p.Jar[u.Host] = cookies
}

// Cookies ...
func (p *CookieJar) Cookies(u *url.URL) []*http.Cookie {
	// fmt.Printf("The URL is : %s\n", u.String())
	// fmt.Printf("Cookie being returned is : %s\n", p.jar[u.Host])
	return p.Jar[u.Host]
}
