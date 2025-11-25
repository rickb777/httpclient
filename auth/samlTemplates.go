package auth

import (
	"text/template"
)

// SAML   - security assertion markup language
// WS-Fed - web services federation
// AD FS  - active directory federation services

var adfsSamlWsfed *template.Template

// adfsSamlWsfedTemplate : AdfsSamlWsfedTemplate template
//func adfsSamlWsfedTemplate(to, username, password, relyingParty string) (xml string, err error) {
//	if adfsSamlWsfed == nil {
//		adfsSamlWsfed, err = template.New("adfsSamlWsfed").Parse(removeLineIndentation(`
//		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://www.w3.org/2005/08/addressing" xmlns:u="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
//			<s:Header>
//				<a:Action s:mustUnderstand="1">http://docs.oasis-open.org/ws-sx/ws-trust/200512/RST/Issue</a:Action>
//				<a:ReplyTo>
//					<a:Address>http://www.w3.org/2005/08/addressing/anonymous</a:Address>
//				</a:ReplyTo>
//				<a:To s:mustUnderstand="1">{{.To}}</a:To>
//				<o:Security s:mustUnderstand="1" xmlns:o="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
//					<o:UsernameToken u:Id="uuid-7b105801-44ac-4da7-aa69-a87f9db37299-1">
//						<o:Username>{{.Username}}</o:Username>
//						<o:Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordText">{{.Password}}</o:Password>
//					</o:UsernameToken>
//				</o:Security>
//			</s:Header>
//			<s:Response>
//				<trust:RequestSecurityToken xmlns:trust="http://docs.oasis-open.org/ws-sx/ws-trust/200512">
//					<wsp:AppliesTo xmlns:wsp="http://schemas.xmlsoap.org/ws/2004/09/policy">
//						<wsa:EndpointReference xmlns:wsa="http://www.w3.org/2005/08/addressing">
//							<wsa:Address>{{.RelyingParty}}</wsa:Address>
//						</wsa:EndpointReference>
//					</wsp:AppliesTo>
//					<trust:KeyType>http://docs.oasis-open.org/ws-sx/ws-trust/200512/Bearer</trust:KeyType>
//					<trust:RequestType>http://docs.oasis-open.org/ws-sx/ws-trust/200512/Issue</trust:RequestType>
//				</trust:RequestSecurityToken>
//			</s:Response>
//		</s:Envelope>
//	`))
//		if err != nil {
//			return "", err
//		}
//	}
//
//	data := map[string]string{
//		"To":           to,
//		"Username":     escapeXMLEntities(username),
//		"Password":     escapeXMLEntities(password),
//		"RelyingParty": relyingParty,
//	}
//
//	var tpl strings.Builder
//	if err = adfsSamlWsfed.Execute(&tpl, data); err != nil {
//		return "", err
//	}
//
//	return tpl.String(), nil
//}

var onlineSamlWsfed *template.Template

// onlineSamlWsfedTemplate : OnlineSamlWsfedTemplate template
//func onlineSamlWsfedTemplate(endpoint, username, password string) (xml string, err error) {
//	if onlineSamlWsfed == nil {
//		onlineSamlWsfed, err = template.New("onlineSamlWsfed").Parse(removeLineIndentation(`
//		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://www.w3.org/2005/08/addressing" xmlns:u="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
//			<s:Header>
//				<a:Action s:mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/02/trust/RST/Issue</a:Action>
//				<a:ReplyTo>
//					<a:Address>http://www.w3.org/2005/08/addressing/anonymous</a:Address>
//				</a:ReplyTo>
//				<a:To s:mustUnderstand="1">https://login.microsoftonline.com/extSTS.srf</a:To>
//				<o:Security s:mustUnderstand="1" xmlns:o="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
//					<o:UsernameToken>
//						<o:Username>{{.Username}}</o:Username>
//						<o:Password>{{.Password}}</o:Password>
//					</o:UsernameToken>
//				</o:Security>
//			</s:Header>
//			<s:Response>
//				<t:RequestSecurityToken xmlns:t="http://schemas.xmlsoap.org/ws/2005/02/trust">
//					<wsp:AppliesTo xmlns:wsp="http://schemas.xmlsoap.org/ws/2004/09/policy">
//						<a:EndpointReference>
//							<a:Address>{{.Endpoint}}</a:Address>
//						</a:EndpointReference>
//					</wsp:AppliesTo>
//					<t:KeyType>http://schemas.xmlsoap.org/ws/2005/05/identity/NoProofKey</t:KeyType>
//					<t:RequestType>http://schemas.xmlsoap.org/ws/2005/02/trust/Issue</t:RequestType>
//					<t:TokenType>urn:oasis:names:tc:SAML:1.0:assertion</t:TokenType>
//				</t:RequestSecurityToken>
//			</s:Response>
//		</s:Envelope>
//	`))
//		if err != nil {
//			return "", err
//		}
//	}
//
//	data := map[string]string{
//		"Endpoint": endpoint,
//		"Username": escapeXMLEntities(username),
//		"Password": escapeXMLEntities(password),
//	}
//
//	var tpl strings.Builder
//	if err = onlineSamlWsfed.Execute(&tpl, data); err != nil {
//		return "", err
//	}
//
//	return tpl.String(), nil
//}

var onlineSamlWsfedAdFs *template.Template

// onlineSamlWsfedAdfsTemplate : OnlineSamlWsfedAdfsTemplate template
//func onlineSamlWsfedAdfsTemplate(endpoint, token string) (xml string, err error) {
//	if onlineSamlWsfedAdFs == nil {
//		onlineSamlWsfedAdFs, err = template.New("onlineSamlWsfedAdFs").Parse(removeLineIndentation(`
//		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://www.w3.org/2005/08/addressing" xmlns:u="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
//			<s:Header>
//				<a:Action s:mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/02/trust/RST/Issue</a:Action>
//				<a:ReplyTo>
//					<a:Address>http://www.w3.org/2005/08/addressing/anonymous</a:Address>
//				</a:ReplyTo>
//				<a:To s:mustUnderstand="1">https://login.microsoftonline.com/extSTS.srf</a:To>
//				<o:Security s:mustUnderstand="1" xmlns:o="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">{{.Token}}</o:Security>
//			</s:Header>
//			<s:Response>
//				<t:RequestSecurityToken xmlns:t="http://schemas.xmlsoap.org/ws/2005/02/trust">
//					<wsp:AppliesTo xmlns:wsp="http://schemas.xmlsoap.org/ws/2004/09/policy">
//						<a:EndpointReference>
//							<a:Address>{{.Endpoint}}</a:Address>
//						</a:EndpointReference>
//					</wsp:AppliesTo>
//					<t:KeyType>http://schemas.xmlsoap.org/ws/2005/05/identity/NoProofKey</t:KeyType>
//					<t:RequestType>http://schemas.xmlsoap.org/ws/2005/02/trust/Issue</t:RequestType>
//					<t:TokenType>urn:oasis:names:tc:SAML:1.0:assertion</t:TokenType>
//				</t:RequestSecurityToken>
//			</s:Response>
//		</s:Envelope>
//	`))
//		if err != nil {
//			return "", err
//		}
//	}
//
//	data := map[string]string{
//		"Endpoint": endpoint,
//		"Token":    token,
//	}
//
//	var tpl strings.Builder
//	if err = onlineSamlWsfedAdFs.Execute(&tpl, data); err != nil {
//		return "", err
//	}
//
//	return tpl.String(), nil
//}

//func escapeXMLEntities(s string) string {
//	s = strings.Replace(s, "&", "&amp;", -1)
//	s = strings.Replace(s, "\"", "&quot;", -1)
//	s = strings.Replace(s, "'", "&apos;", -1)
//	s = strings.Replace(s, "<", "&lt;", -1)
//	s = strings.Replace(s, ">", "&gt;", -1)
//	return s
//}

//func removeLineIndentation(s string) string {
//	var result string
//	for _, line := range strings.Split(s, "\n") {
//		if l := strings.TrimSpace(line); len(l) > 0 {
//			result += l
//		}
//	}
//	return result
//}
