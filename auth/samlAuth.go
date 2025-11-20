package auth

// SAML provides authentication using Security Authentication Markup Language.
// See https://tools.ietf.org/html/rfc7522
// TODO finish and test this
//func SAML(user, pw, siteURL string, hc httpclient.HttpClient) Authenticator {
//	return &samlAuth{
//		user:    user,
//		pw:      pw,
//		siteURL: siteURL,
//		hc:      hc,
//	}
//}

//type samlAuth struct {
//	user    string
//	pw      string
//	siteURL string
//	hc      httpclient.HttpClient
//}

// Type identifies the Basic authenticator.
//func (sa *samlAuth) Type() string {
//	return "SAML"
//}

// User holds the BasicAuth username.
//func (sa *samlAuth) User() string {
//	return sa.user
//}

// pw holds the BasicAuth password.
//func (sa *samlAuth) Password() string {
//	return sa.pw
//}

//func (sa *samlAuth) Challenge([]string) Authenticator {
//	return sa
//}

// Authorize the current request.
//func (sa *samlAuth) Authorize(req *http.Request) {
//	authCookie, _, err := sa.getAuth()
//	if err == nil {
//		req.Header.Set("Cookie", authCookie)
//	}
//}

//func (sa *samlAuth) post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
//	req, err := http.NewRequest("POST", url, body)
//	if err != nil {
//		return nil, err
//	}
//	req.Header.Set("Content-Type", contentType)
//	return sa.hc.Do(req)
//}

//func (sa *samlAuth) getAuth() (string, int64, error) {
//	if sa.hc == nil {
//		sa.hc = http.DefaultClient
//	}
//
//	parsedURL, err := url.Parse(sa.siteURL)
//	if err != nil {
//		return "", 0, err
//	}
//
//	cacheKey := parsedURL.Host + "@saml@" + sa.user + "@" + sa.pw
//	if authToken, exp, found := storage.GetWithExpiration(cacheKey); found {
//		return authToken.(string), exp.Unix(), nil
//	}
//
//	authCookie, notAfter, err := getSecurityToken(sa)
//	if err != nil {
//		return "", 0, err
//	}
//
//	notAfterTime, _ := time.Parse(time.RFC3339, notAfter)
//	expiry := time.Until(notAfterTime) - 60*time.Second
//	exp := time.Now().Add(expiry).Unix()
//
//	storage.Set(cacheKey, authCookie, expiry)
//
//	return authCookie, exp, nil
//	return "", 0, nil
//}

//func getSecurityToken(sa *samlAuth) (string, string, error) {
//	if sa.hc == nil {
//		sa.hc = http.DefaultClient
//	}
//
//	const endpoint = "https://login.microsoftonline.com/GetUserRealm.srf"
//
//	params := url.Values{}
//	params.Set("login", sa.user)
//
//	sa.hc.(*http.Client).CheckRedirect = doNotCheckRedirect
//
//	resp, err := sa.post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
//	if err != nil {
//		return "", "", err
//	}
//
//	defer func() {
//		if resp != nil && resp.Body != nil {
//			_ = resp.Body.Close()
//		}
//	}()
//
//	data, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", "", err
//	}
//
//	type userReadlmResponse struct {
//		NameSpaceType       string `json:"NameSpaceType"`
//		DomainName          string `json:"DomainName"`
//		FederationBrandName string `json:"FederationBrandName"`
//		CloudInstanceName   string `json:"CloudInstanceName"`
//		State               int    `json:"State"`
//		UserState           int    `json:"UserState"`
//		Login               string `json:"Login"`
//		AuthURL             string `json:"AuthURL"`
//	}
//
//	userRealm := &userReadlmResponse{}
//	if err = json.Unmarshal(data, &userRealm); err != nil {
//		return "", "", err
//	}
//
//	// fmt.Printf("Results: %v\n", userRealm.NameSpaceType)
//
//	if userRealm.NameSpaceType == "" {
//		return "", "", errors.New("unable to define namespace type for Online authentiation")
//	}
//
//	if userRealm.NameSpaceType == "Managed" {
//		return getSecurityTokenWithOnline(sa)
//	}
//
//	if userRealm.NameSpaceType == "Federated" {
//		return getSecurityTokenWithAdfs(userRealm.AuthURL, sa)
//	}
//
//	return "", "", fmt.Errorf("unable to resolve namespace authentiation type. Type received: %s", userRealm.NameSpaceType)
//}

//func getSecurityTokenWithOnline(sa *samlAuth) (string, string, error) {
//	if sa.hc == nil {
//		sa.hc = http.DefaultClient
//	}
//
//	parsedURL, err := url.Parse(sa.siteURL)
//	if err != nil {
//		return "", "", err
//	}
//
//	formsEndpoint := fmt.Sprintf("%s://%s/_forms/default.aspx?wa=wsignin1.0", parsedURL.Scheme, parsedURL.Host)
//	samlBody, err := onlineSamlWsfedTemplate(formsEndpoint, sa.user, sa.pw)
//	if err != nil {
//		return "", "", err
//	}
//
//	stsEndpoint := "https://login.microsoftonline.com/extSTS.srf" // TODO: add mapping for diff SPOs
//
//	req, err := http.NewRequest("POST", stsEndpoint, bytes.NewBuffer([]byte(samlBody)))
//	if err != nil {
//		return "", "", err
//	}
//
//	req.Header.Set("Content-Type", "application/soap+xml;charset=utf-8")
//
//	sa.hc.(*http.Client).CheckRedirect = doNotCheckRedirect
//
//	resp, err := sa.hc.Do(req)
//	if err != nil {
//		return "", "", err
//	}
//	defer func() {
//		if resp != nil && resp.Body != nil {
//			_ = resp.Body.Close()
//		}
//	}()
//
//	xmlResponse, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", "", err
//	}
//
//	type samlAssertion struct {
//		Fault    string `xml:"Body>Fault>Reason>Text"`
//		Response struct {
//			BinaryToken string `xml:"RequestedSecurityToken>BinarySecurityToken"`
//			Lifetime    struct {
//				Created string `xml:"Created"`
//				Expires string `xml:"Expires"`
//			} `xml:"Lifetime"`
//		} `xml:"Body>RequestSecurityTokenResponse"`
//	}
//	result := &samlAssertion{}
//	if err := xml.Unmarshal(xmlResponse, &result); err != nil {
//		return "", "", err
//	}
//
//	resp, err = sa.post(formsEndpoint, "application/x-www-form-urlencoded", strings.NewReader(result.Response.BinaryToken))
//	if err != nil {
//		return "", "", err
//	}
//	defer func() {
//		if resp != nil && resp.Body != nil {
//			_ = resp.Body.Close()
//		}
//	}()
//
//	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
//		return "", "", err
//	}
//
//	// cookie := resp.Header.Get("Set-Cookie") // TODO: parse FedAuth and rtFa cookies only (?)
//	// fmt.Printf("Cookie: %s\n", cookie)
//	// fmt.Printf("Resp2, %v\n", resp.StatusCode)
//
//	var authCookie string
//	for _, coo := range resp.Cookies() {
//		if coo.Name == "rtFa" || coo.Name == "FedAuth" {
//			authCookie += coo.String() + "; "
//		}
//	}
//
//	return authCookie, result.Response.Lifetime.Expires, nil
//}

//func getSecurityTokenWithAdfs(adfsURL string, sa *samlAuth) (string, string, error) {
//	if sa.hc == nil {
//		sa.hc = http.DefaultClient
//	}
//
//	parsedAdfsURL, err := url.Parse(adfsURL)
//	if err != nil {
//		return "", "", err
//	}
//
//	usernameMixedURL := fmt.Sprintf("%s://%s/adfs/services/trust/13/usernamemixed", parsedAdfsURL.Scheme, parsedAdfsURL.Host)
//	samlBody, err := adfsSamlWsfedTemplate(usernameMixedURL, sa.user, sa.pw, "urn:federation:MicrosoftOnline")
//	if err != nil {
//		return "", "", err
//	}
//
//	sa.hc.(*http.Client).CheckRedirect = doNotCheckRedirect
//	resp, err := sa.post(usernameMixedURL, "application/soap+xml;charset=utf-8", bytes.NewBuffer([]byte(samlBody)))
//	if err != nil {
//		return "", "", err
//	}
//	defer func() {
//		if resp != nil && resp.Body != nil {
//			_ = resp.Body.Close()
//		}
//	}()
//
//	res, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", "", err
//	}
//
//	type samlAssertion struct {
//		Response struct {
//			Fault string `xml:"Fault>Reason>Text"`
//			Token struct {
//				Inner      []byte `xml:",innerxml"`
//				Conditions struct {
//					NotBefore    string `xml:"NotBefore,attr"`
//					NotOnOrAfter string `xml:"NotOnOrAfter,attr"`
//				} `xml:"Assertion>Conditions"`
//			} `xml:"RequestSecurityTokenResponseCollection>RequestSecurityTokenResponse>RequestedSecurityToken"`
//		} `xml:"Body"`
//	}
//
//	result := &samlAssertion{}
//	if err := xml.Unmarshal(res, &result); err != nil {
//		return "", "", err
//	}
//
//	if result.Response.Fault != "" {
//		return "", "", errors.New(result.Response.Fault)
//	}
//
//	parsedURL, err := url.Parse(sa.siteURL)
//	if err != nil {
//		return "", "", err
//	}
//
//	rootSite := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
//	tokenRequest, err := onlineSamlWsfedAdfsTemplate(rootSite, string(result.Response.Token.Inner))
//	if err != nil {
//		return "", "", err
//	}
//
//	stsEndpoint := "https://login.microsoftonline.com/extSTS.srf" // TODO: mapping
//
//	resp, err = sa.post(stsEndpoint, "application/soap+xml;charset=utf-8", bytes.NewBuffer([]byte(tokenRequest)))
//	if err != nil {
//		return "", "", err
//	}
//	defer func() {
//		if resp != nil && resp.Body != nil {
//			_ = resp.Body.Close()
//		}
//	}()
//
//	xmlResponse, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", "", err
//	}
//
//	type tokenAssertion struct {
//		Fault    string `xml:"Body>Fault>Reason>Text"`
//		Response struct {
//			BinaryToken string `xml:"RequestedSecurityToken>BinarySecurityToken"`
//			Lifetime    struct {
//				Created string `xml:"Created"`
//				Expires string `xml:"Expires"`
//			} `xml:"Lifetime"`
//		} `xml:"Body>RequestSecurityTokenResponse"`
//	}
//
//	tokenResult := &tokenAssertion{}
//	if err := xml.Unmarshal(xmlResponse, &tokenResult); err != nil {
//		return "", "", err
//	}
//
//	if tokenResult.Response.BinaryToken == "" {
//		return "", "", errors.New("can't extract binary token")
//	}
//
//	sa.hc.(*http.Client).CheckRedirect = doNotCheckRedirect
//
//	formsEndpoint := fmt.Sprintf("%s://%s/_forms/default.aspx?wa=wsignin1.0", parsedURL.Scheme, parsedURL.Host)
//	resp, err = sa.post(formsEndpoint, "application/x-www-form-urlencoded", strings.NewReader(tokenResult.Response.BinaryToken))
//	if err != nil {
//		return "", "", err
//	}
//	defer func() {
//		if resp != nil && resp.Body != nil {
//			_ = resp.Body.Close()
//		}
//	}()
//
//	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
//		return "", "", err
//	}
//
//	var authCookie string
//	for _, coo := range resp.Cookies() {
//		if coo.Name == "rtFa" || coo.Name == "FedAuth" {
//			authCookie += coo.String() + "; "
//		}
//	}
//
//	return authCookie, tokenResult.Response.Lifetime.Expires, nil
//}

// doNotCheckRedirect *http.Client CheckRedirect callback to ignore redirects
//func doNotCheckRedirect(_ *http.Request, _ []*http.Request) error {
//	return http.ErrUseLastResponse
//}

//var (
//	storage = cache.New(5*time.Minute, 10*time.Minute)
//)
