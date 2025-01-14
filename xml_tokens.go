package fastxml

import (
	"bytes"
)

type XMLEscapingMode int

const (
	NoEscaping XMLEscapingMode = iota
	XMLEscapeMode
	XMLUnescapeMode
)

type xmlTokenType int

const (
	unknownXMLToken    xmlTokenType = iota //unknown token
	startXMLToken                          //<ns:xmltag k1="v1" k2='v2'>
	inlineXMLToken                         //<ns:xmltag/>
	endXMLToken                            //</ns:xmltag>
	processingXMLToken                     //<? text ?>
	commentsXMLToken                       //<!-- text -->
	cdataXMLToken                          //<![CDATA[ text ]]>
	doctypeXMLToken                        //<!DOCTYPE [ text ]>
	textToken                              //text

	maxXMLTokenType
)

var (
	cdataStart = []byte("<![CDATA[")
	cdataEnd   = []byte("]]>")
)

func (t xmlTokenType) String() string {
	return []string{
		"UnknownTokenType",
		"StartXMLToken",
		"InlineXMLToken",
		"EndXMLToken",
		"ProcessingXMLToken",
		"CommentsXMLToken",
		"CDATAToken",
		"DOCTYPEToken",
		"TextToken",
	}[t]
}

type xmlTagIndex struct {
	si, ei int
}

type XMLToken struct {
	start, end xmlTagIndex
	name, text xmlTagIndex
	cdata      bool
}

func NewXMLToken(ssi, sei, esi, eei int) XMLToken {
	return XMLToken{
		start: xmlTagIndex{si: ssi, ei: sei},
		end:   xmlTagIndex{si: esi, ei: eei},
	}
}

func (t *XMLToken) XMLTag(in []byte) []byte {
	return in[t.start.si:t.end.ei]
}

func (t *XMLToken) Text(in []byte) []byte {
	if t.start.si == t.end.si {
		return nil //inline tag doesn't have text
	}
	if t.text.si == 0 {
		t.text.si, t.text.ei, t.cdata = _trimCDATA(in, t.start.ei, t.end.si)
	}
	return in[t.text.si:t.text.ei]
}

func (t *XMLToken) Name(in []byte) []byte {
	if t.name.si == 0 {
		t.name.si, t.name.ei = getTokenNameIndex(in, t.start.si+1)
	}
	return in[t.name.si:t.name.ei]
}

func (t *XMLToken) NSName(in []byte) []byte {
	if t.name.si == 0 {
		t.name.si, t.name.ei = getTokenNameIndex(in, t.start.si+1)
	}
	return in[t.start.si+1 : t.name.ei]
}

func (t XMLToken) ParseAttribute(in []byte) []Attribute {
	offset := 1
	if t.start.si == t.end.si {
		offset = 2 //check for inline token eg: <test k="v"/>
	}
	return parseAttributes(in[:], t.name.ei, t.start.ei-offset)
}

func (t XMLToken) IsInline() bool {
	return (t.start.ei == t.end.ei)
}

func (t XMLToken) StartTagOffset() (si, ei int) {
	return t.start.si, t.start.ei
}

func (t XMLToken) EndTagOffset() (si, ei int) {
	return t.end.si, t.end.ei
}

func (t XMLToken) TagOffset() (si, ei int) {
	return t.start.si, t.end.ei
}

func (t XMLToken) IsCDATA(in []byte) bool {
	if t.start.si == t.end.si {
		return false //inline tag doesn't have text
	}
	if t.text.si == 0 {
		t.text.si, t.text.ei, t.cdata = _trimCDATA(in, t.start.ei, t.end.si)
	}
	return t.cdata
}

func getTokenType(in []byte, index int) xmlTokenType {
	if index >= len(in) {
		//TODO: donot check for negative values,
		//as this is used by internally and this condition won't happens
		return unknownXMLToken
	}
	//TODO: check if removing whitespace required
	ch := in[index]
	switch ch {
	case '/':
		return endXMLToken
	case '!':
		//TODO: check if removing whitespace required
		index++
		if index >= len(in) {
			return unknownXMLToken
		}
		ch1 := in[index]
		switch ch1 {
		case '-':
			if index+1 < len(in) && in[index+1] == '-' {
				//check for full comment <!--
				return commentsXMLToken
			}
		case '[':
			if bytes.HasPrefix(in[index:], []byte(`[CDATA[`)) {
				//check for full <![CDATA[
				return cdataXMLToken
			}
		case 'D':
			if bytes.HasPrefix(in[index:], []byte(`DOCTYPE`)) {
				//check for full <!DOCTYPE
				return doctypeXMLToken
			}
		}
	case '?':
		return processingXMLToken
	default:
		if alpha[ch] {
			return startXMLToken
		}
	}
	return unknownXMLToken
}

func getTokenNameIndex(in []byte, startIndex int) (si, ei int) {
	var firstNameSpace bool
	si = startIndex
	for i := si; i < len(in); i++ {
		if name[in[i]] {
			continue
		}
		if in[i] == '>' || whitespace[in[i]] || in[i] == '/' {
			if alpha[in[si]] || in[si] == '_' {
				return si, i
			}
		} else if in[i] == ':' && !firstNameSpace {
			//TODO: return namespace indexes too
			si = i + 1
			firstNameSpace = true
			continue
		}
		break //invalid token name
	}
	return startIndex, startIndex //not found
}

func getTokenEndIndex(in []byte, startIndex int, ttype xmlTokenType) (int, bool) {
	index := -1
	inline := false
	//TODO: write token type based parsers and execute it separately
	switch ttype {
	case startXMLToken:
		// read until >
		for i := startIndex; i < len(in); i++ {
			if in[i] == '>' {
				if i > 0 && in[i-1] == '/' {
					inline = true
				}
				//found end tag
				index = i + 1
				break
			}
		}
	case endXMLToken:
		// read until >
		for i := startIndex; i < len(in); i++ {
			if in[i] == '>' {
				//found end tag
				index = i + 1
				break
			}
		}
	case processingXMLToken:
		// read until ?>
		for i := startIndex; i < len(in); i++ {
			if in[i] == '>' && i > 0 && in[i-1] == '?' {
				//found end tag
				index = i + 1
				break
			}
		}
	case commentsXMLToken:
		// read until found -->
		for i := startIndex; i < len(in); i++ {
			if in[i] == '>' && i > 1 && in[i-1] == '-' && in[i-2] == '-' {
				//found end tag
				index = i + 1
				break
			}
		}
	case cdataXMLToken:
		// read until ]]> /*<![CDATA[ 25.00 ]]>*/
		for i := startIndex; i < len(in); i++ {
			if in[i] == '>' && i > 1 && in[i-1] == ']' && in[i-2] == ']' {
				/*
					TODO: Special handling (https://en.wikipedia.org/wiki/CDATA#Nesting)
					input: <![CDATA[ data ]]> data ]]>
					replace ]]> with ]]]]><![CDATA[>
					output: <![CDATA[ data ]]]]><![CDATA[> data ]]>
					action: ignore if found ']]]]><![CDATA[>'
				*/
				//found end tag
				index = i + 1
				break
			}
		}
	case doctypeXMLToken:
		/*
			TODO:
			1. UNDER DEVELOPMENT
			2. Nested DCOTYPE strict checking not supported yet
			if DOCTYPE contains [ start then ends with ]> else ends with >
		*/
		var bracketCounts int
		for i := startIndex; i < len(in); i++ {
			if in[i] == '>' {
				if bracketCounts == 0 {
					//found end tag
					index = i + 1
				} else {
					if bracketCounts == 1 && i > 0 && in[i-1] == ']' {
						index = i + 1
					} else {
						bracketCounts--
						continue
					}
				}
				break
			} else if in[i] == '[' {
				//TODO: Buggy no strict checking of tag
				bracketCounts++
			}
		}
	default:
		//read token based on tokentype
		for i := startIndex; i < len(in); i++ {
			if in[i] == '>' {
				index = i + 1
				break
			}
		}
	}
	return index, inline
}
