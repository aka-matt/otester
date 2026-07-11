package security

import (
	"regexp"
	"strings"
)

var (
	clientSecretPattern = regexp.MustCompile(`(client_secret=)([^&]+)`)
	accessTokenPattern  = regexp.MustCompile(`(access_token=)([^&]+)`)
	clientSecretJSON    = regexp.MustCompile(`"(client_secret)"\s*:\s*"[^"]+"`)
	accessTokenJSON     = regexp.MustCompile(`"(access_token)"\s*:\s*"[^"]+"`)
	bearerPattern       = regexp.MustCompile(`(?i)(Bearer )([^\s]+)`)
)

func RedactURL(url string) string {
	url = clientSecretPattern.ReplaceAllString(url, "$1******")
	url = accessTokenPattern.ReplaceAllString(url, "$1******")
	return url
}

func RedactBearerToken(token string) string {
	if len(token) <= 10 {
		return "******"
	}
	return token[:6] + "..." + token[len(token)-4:]
}

func RedactAuthorizationHeader(header string) string {
	return bearerPattern.ReplaceAllStringFunc(header, func(match string) string {
		parts := bearerPattern.FindStringSubmatch(match)
		if len(parts) >= 3 {
			return parts[1] + RedactBearerToken(parts[2])
		}
		return "******"
	})
}

func RedactHeaders(headers map[string][]string) map[string][]string {
	redacted := make(map[string][]string)
	for k, v := range headers {
		if strings.EqualFold(k, "authorization") {
			redacted[k] = []string{RedactAuthorizationHeader(v[0])}
		} else {
			redacted[k] = v
		}
	}
	return redacted
}

func RedactJSON(jsonStr string) string {
	jsonStr = clientSecretJSON.ReplaceAllString(jsonStr, `"$1":"******"`)
	jsonStr = accessTokenJSON.ReplaceAllString(jsonStr, `"$1":"******"`)
	return jsonStr
}
