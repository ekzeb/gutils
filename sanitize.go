package util


import (
"bytes"
"html"
"html/template"
"io"
"path"
"regexp"
"strings"

parser "golang.org/x/net/html"
)

var (
	ignoreTags = []string{"title", "script", "style", "iframe", "frame", "frameset", "noframes", "noembed", "embed", "applet", "object", "base"}

	defaultTags = []string{"h1", "h2", "h3", "h4", "h5", "h6", "div", "span", "hr", "p", "br", "b", "i", "strong", "em", "ol", "ul", "li", "a", "img", "pre", "code", "blockquote"}

	defaultAttributes = []string{"id", "class", "src", "href", "title", "alt", "name", "rel"}
)

// HTMLAllowing sanitizes html, allowing some tags.
// Arrays of allowed tags and allowed attributes may optionally be passed as the second and third arguments.
func HTMLAllowing(s string, args ...[]string) (string, error) {

	allowedTags := defaultTags
	if len(args) > 0 {
		allowedTags = args[0]
	}
	allowedAttributes := defaultAttributes
	if len(args) > 1 {
		allowedAttributes = args[1]
	}

	// Parse the html
	tokenizer := parser.NewTokenizer(strings.NewReader(s))

	buffer := bytes.NewBufferString("")
	ignore := ""

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		switch tokenType {

		case parser.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				return buffer.String(), nil
			}
			return "", err

		case parser.StartTagToken:

			if len(ignore) == 0 && includes(allowedTags, token.Data) {
				token.Attr = cleanAttributes(token.Attr, allowedAttributes)
				buffer.WriteString(token.String())
			} else if includes(ignoreTags, token.Data) {
				ignore = token.Data
			}

		case parser.SelfClosingTagToken:

			if len(ignore) == 0 && includes(allowedTags, token.Data) {
				token.Attr = cleanAttributes(token.Attr, allowedAttributes)
				buffer.WriteString(token.String())
			} else if token.Data == ignore {
				ignore = ""
			}

		case parser.EndTagToken:
			if len(ignore) == 0 && includes(allowedTags, token.Data) {
				token.Attr = []parser.Attribute{}
				buffer.WriteString(token.String())
			} else if token.Data == ignore {
				ignore = ""
			}

		case parser.TextToken:
			// We allow text content through, unless ignoring this entire tag and its contents (including other tags)
			if ignore == "" {
				buffer.WriteString(token.String())
			}
		case parser.CommentToken:
		// We ignore comments by default
		case parser.DoctypeToken:
		// We ignore doctypes by default - html5 does not require them and this is intended for sanitizing snippets of text
		default:
		// We ignore unknown token types by default

		}

	}

}

// HTML strips html tags, replace common entities, and escapes <>&;'" in the result.
// Note the returned text may contain entities as it is escaped by HTMLEscapeString, and most entities are not translated.
func HTML(s string) (output string) {

	// Shortcut strings with no tags in them
	if !strings.ContainsAny(s, "<>") {
		output = s
	} else {

		// First remove line breaks etc as these have no meaning outside html tags (except pre)
		// this means pre sections will lose formatting... but will result in less unintentional paras.
		s = strings.Replace(s, "\n", "", -1)

		// Then replace line breaks with newlines, to preserve that formatting
		s = strings.Replace(s, "</p>", "\n", -1)
		s = strings.Replace(s, "<br>", "\n", -1)
		s = strings.Replace(s, "</br>", "\n", -1)
		s = strings.Replace(s, "<br/>", "\n", -1)
		s = strings.Replace(s, "<br />", "\n", -1)

		// Walk through the string removing all tags
		b := bytes.NewBufferString("")
		inTag := false
		for _, r := range s {
			switch r {
			case '<':
				inTag = true
			case '>':
				inTag = false
			default:
				if !inTag {
					b.WriteRune(r)
				}
			}
		}
		output = b.String()
	}

	// Remove a few common harmless entities, to arrive at something more like plain text
	output = strings.Replace(output, "&#8216;", "'", -1)
	output = strings.Replace(output, "&#8217;", "'", -1)
	output = strings.Replace(output, "&#8220;", "\"", -1)
	output = strings.Replace(output, "&#8221;", "\"", -1)
	output = strings.Replace(output, "&nbsp;", " ", -1)
	output = strings.Replace(output, "&quot;", "\"", -1)
	output = strings.Replace(output, "&apos;", "'", -1)

	// Translate some entities into their plain text equivalent (for example accents, if encoded as entities)
	output = html.UnescapeString(output)

	// In case we have missed any tags above, escape the text - removes <, >, &, ' and ".
	output = template.HTMLEscapeString(output)

	// After processing, remove some harmless entities &, ' and " which are encoded by HTMLEscapeString
	output = strings.Replace(output, "&#34;", "\"", -1)
	output = strings.Replace(output, "&#39;", "'", -1)
	output = strings.Replace(output, "&amp; ", "& ", -1)     // NB space after
	output = strings.Replace(output, "&amp;amp; ", "& ", -1) // NB space after

	return output
}

// We are very restrictive as this is intended for ascii url slugs
var illegalPath = regexp.MustCompile(`[^[:alnum:]\~\-\./]`)

// Path makes a string safe to use as an url path.
func Path(s string) string {
	// Start with lowercase string
	filePath := strings.ToLower(s)
	filePath = strings.Replace(filePath, "..", "", -1)
	filePath = path.Clean(filePath)

	// Remove illegal characters for paths, flattening accents and replacing some common separators with -
	filePath = cleanString(filePath, illegalPath)

	// NB this may be of length 0, caller must check
	return filePath
}

// Remove all other unrecognised characters apart from
var illegalName = regexp.MustCompile(`[^[:alnum:]-.]`)

// Name makes a string safe to use in a file name by first finding the path basename, then replacing non-ascii characters.
func Name(s string) string {
	// Start with lowercase string
	fileName := strings.ToLower(s)
	fileName = path.Clean(path.Base(fileName))

	// Remove illegal characters for names, replacing some common separators with -
	fileName = cleanString(fileName, illegalName)

	// NB this may be of length 0, caller must check
	return fileName
}

// Replace these separators with -
var baseNameSeparators = regexp.MustCompile(`[./]`)

// BaseName makes a string safe to use in a file name, producing a sanitized basename replacing . or / with -.
// No attempt is made to normalise a path or normalise case.
func BaseName(s string) string {

	// Replace certain joining characters with a dash
	baseName := baseNameSeparators.ReplaceAllString(s, "-")

	// Remove illegal characters for names, replacing some common separators with -
	baseName = cleanString(baseName, illegalName)

	// NB this may be of length 0, caller must check
	return baseName
}

// A very limited list of transliterations to catch common european names translated to urls.
// This set could be expanded with at least caps and many more characters.
var transliterations = map[rune]string{
	'À': "A",
	'Á': "A",
	'Â': "A",
	'Ã': "A",
	'Ä': "A",
	'Å': "AA",
	'Æ': "AE",
	'Ç': "C",
	'È': "E",
	'É': "E",
	'Ê': "E",
	'Ë': "E",
	'Ì': "I",
	'Í': "I",
	'Î': "I",
	'Ï': "I",
	'Ð': "D",
	'Ł': "L",
	'Ñ': "N",
	'Ò': "O",
	'Ó': "O",
	'Ô': "O",
	'Õ': "O",
	'Ö': "O",
	'Ø': "OE",
	'Ù': "U",
	'Ú': "U",
	'Ü': "U",
	'Û': "U",
	'Ý': "Y",
	'Þ': "Th",
	'ß': "ss",
	'à': "a",
	'á': "a",
	'â': "a",
	'ã': "a",
	'ä': "a",
	'å': "aa",
	'æ': "ae",
	'ç': "c",
	'è': "e",
	'é': "e",
	'ê': "e",
	'ë': "e",
	'ì': "i",
	'í': "i",
	'î': "i",
	'ï': "i",
	'ð': "d",
	'ł': "l",
	'ñ': "n",
	'ń': "n",
	'ò': "o",
	'ó': "o",
	'ô': "o",
	'õ': "o",
	'ō': "o",
	'ö': "o",
	'ø': "oe",
	'ś': "s",
	'ù': "u",
	'ú': "u",
	'û': "u",
	'ū': "u",
	'ü': "u",
	'ý': "y",
	'þ': "th",
	'ÿ': "y",
	'ż': "z",
	'Œ': "OE",
	'œ': "oe",
	/* x004 */
	0x0400: "Ie",
	0x0401: "Io",
	0x0402: "Dj",
	0x0403: "Gj",
	0x0404: "Ie",
	0x0405: "Dz",
	0x0406: "I",
	0x0407: "Yi",
	0x0408: "J",
	0x0409: "Lj",
	0x040a: "Nj",
	0x040b: "Tsh",
	0x040c: "Kj",
	0x040d: "I",
	0x040e: "U",
	0x040f: "Dzh",
	0x0410: "A",
	0x0411: "B",
	0x0412: "V",
	0x0413: "G",
	0x0414: "D",
	0x0415: "E",
	0x0416: "Zh",
	0x0417: "Z",
	0x0418: "I",
	0x0419: "I",
	0x041a: "K",
	0x041b: "L",
	0x041c: "M",
	0x041d: "N",
	0x041e: "O",
	0x041f: "P",
	0x0420: "R",
	0x0421: "S",
	0x0422: "T",
	0x0423: "U",
	0x0424: "F",
	0x0425: "Kh",
	0x0426: "Ts",
	0x0427: "Ch",
	0x0428: "Sh",
	0x0429: "Shch",
	0x042a: "",
	0x042b: "Y",
	0x042c: "",
	0x042d: "E",
	0x042e: "Iu",
	0x042f: "Ia",
	0x0430: "a",
	0x0431: "b",
	0x0432: "v",
	0x0433: "g",
	0x0434: "d",
	0x0435: "e",
	0x0436: "zh",
	0x0437: "z",
	0x0438: "i",
	0x0439: "i",
	0x043a: "k",
	0x043b: "l",
	0x043c: "m",
	0x043d: "n",
	0x043e: "o",
	0x043f: "p",
	0x0440: "r",
	0x0441: "s",
	0x0442: "t",
	0x0443: "u",
	0x0444: "f",
	0x0445: "kh",
	0x0446: "ts",
	0x0447: "ch",
	0x0448: "sh",
	0x0449: "shch",
	0x044a: "",
	0x044b: "y",
	0x044c: "",
	0x044d: "e",
	0x044e: "iu",
	0x044f: "ia",
	0x0450: "ie",
	0x0451: "io",
	0x0452: "dj",
	0x0453: "gj",
	0x0454: "ie",
	0x0455: "dz",
	0x0456: "i",
	0x0457: "yi",
	0x0458: "j",
	0x0459: "lj",
	0x045a: "nj",
	0x045b: "tsh",
	0x045c: "kj",
	0x045d: "i",
	0x045e: "u",
	0x045f: "dzh",
	0x0460: "O",
	0x0461: "o",
	0x0462: "E",
	0x0463: "e",
	0x0464: "Ie",
	0x0465: "ie",
	0x0466: "E",
	0x0467: "e",
	0x0468: "Ie",
	0x0469: "ie",
	0x046a: "O",
	0x046b: "o",
	0x046c: "Io",
	0x046d: "io",
	0x046e: "Ks",
	0x046f: "ks",
	0x0470: "Ps",
	0x0471: "ps",
	0x0472: "F",
	0x0473: "f",
	0x0474: "Y",
	0x0475: "y",
	0x0476: "Y",
	0x0477: "y",
	0x0478: "u",
	0x0479: "u",
	0x047a: "O",
	0x047b: "o",
	0x047c: "O",
	0x047d: "o",
	0x047e: "Ot",
	0x047f: "ot",
	0x0480: "Q",
	0x0481: "q",
	0x0482: "1000",
	0x0483: "",
	0x0484: "",
	0x0485: "",
	0x0486: "",
	0x0487: "",
	0x0488: "100000",
	0x0489: "1000000",
	0x048a: "",
	0x048b: "",
	0x048c: "",
	0x048d: "",
	0x04ae: "U",
	0x04af: "u",
	0x04b4: "Tts",
	0x04b5: "tts",
	0x04ba: "H",
	0x04bb: "h",
	0x04bc: "Ch",
	0x04bd: "ch",
	0x04c1: "Zh",
	0x04c2: "zh",
	0x04cb: "Ch",
	0x04cc: "ch",
	0x04d0: "a",
	0x04d1: "a",
	0x04d2: "A",
	0x04d3: "a",
	0x04d4: "Ae",
	0x04d5: "ae",
	0x04d6: "Ie",
	0x04d7: "ie",
	0x04dc: "Zh",
	0x04dd: "zh",
	0x04de: "Z",
	0x04df: "z",
	0x04e0: "Dz",
	0x04e1: "dz",
	0x04e2: "I",
	0x04e3: "i",
	0x04e4: "I",
	0x04e5: "i",
	0x04e6: "O",
	0x04e7: "o",
	0x04e8: "O",
	0x04e9: "o",
	0x04ea: "O",
	0x04eb: "o",
	0x04ec: "E",
	0x04ed: "e",
	0x04ee: "U",
	0x04ef: "u",
	0x04f0: "U",
	0x04f1: "u",
	0x04f2: "U",
	0x04f3: "u",
	0x04f4: "Ch",
	0x04f5: "ch",
	0x04f8: "Y",
	0x04f9: "y",
}

// Accents replaces a set of accented characters with ascii equivalents.
func Accents(s string) string {
	// Replace some common accent characters
	b := bytes.NewBufferString("")
	for _, c := range s {
		// Check transliterations first
		if val, ok := transliterations[c]; ok {
			b.WriteString(val)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

var (
	// If the attribute contains data: or javascript: anywhere, ignore it
	// we don't allow this in attributes as it is so frequently used for xss
	// NB we allow spaces in the value, and lowercase.
	illegalAttr = regexp.MustCompile(`(d\s*a\s*t\s*a|j\s*a\s*v\s*a\s*s\s*c\s*r\s*i\s*p\s*t\s*)\s*:`)

	// We are far more restrictive with href attributes.
	legalHrefAttr = regexp.MustCompile(`\A[/#][^/\\]?|mailto://|http://|https://`)
)

// cleanAttributes returns an array of attributes after removing malicious ones.
func cleanAttributes(a []parser.Attribute, allowed []string) []parser.Attribute {
	if len(a) == 0 {
		return a
	}

	var cleaned []parser.Attribute
	for _, attr := range a {
		if includes(allowed, attr.Key) {

			val := strings.ToLower(attr.Val)

			// Check for illegal attribute values
			if illegalAttr.FindString(val) != "" {
				attr.Val = ""
			}

			// Check for legal href values - / mailto:// http:// or https://
			if attr.Key == "href" {
				if legalHrefAttr.FindString(val) == "" {
					attr.Val = ""
				}
			}

			// If we still have an attribute, append it to the array
			if attr.Val != "" {
				cleaned = append(cleaned, attr)
			}
		}
	}
	return cleaned
}

// A list of characters we consider separators in normal strings and replace with our canonical separator - rather than removing.
var (
	separators = regexp.MustCompile(`[ &_=+:]`)

	dashes = regexp.MustCompile(`[\-]+`)
)

// cleanString replaces separators with - and removes characters listed in the regexp provided from string.
// Accents, spaces, and all characters not in A-Za-z0-9 are replaced.
func cleanString(s string, r *regexp.Regexp) string {

	// Remove any trailing space to avoid ending on -
	s = strings.Trim(s, " ")

	// Flatten accents first so that if we remove non-ascii we still get a legible name
	s = Accents(s)

	// Replace certain joining characters with a dash
	s = separators.ReplaceAllString(s, "-")

	// Remove all other unrecognised characters - NB we do allow any printable characters
	s = r.ReplaceAllString(s, "")

	// Remove any multiple dashes caused by replacements above
	s = dashes.ReplaceAllString(s, "-")

	return s
}

// includes checks for inclusion of a string in a []string.
func includes(a []string, s string) bool {
	for _, as := range a {
		if as == s {
			return true
		}
	}
	return false
}
