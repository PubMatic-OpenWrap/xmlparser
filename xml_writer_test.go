package fastxml

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_XMLElement(t *testing.T) {
	tests := []struct {
		name  string
		setup func() XMLWriter
		want  string
	}{
		{
			name: `empty_node`,
			setup: func() XMLWriter {
				return CreateElement("node")
			},
			want: `<node></node>`,
		},
		{
			name: `adding attributes`,
			setup: func() XMLWriter {
				node := CreateElement("node")
				node.AddAttribute("ns", "key1", "value1")
				node.AddAttribute("", "key2", "value2")
				node.SetText("text", false, NoEscaping)
				return node
			},
			want: `<node ns:key1="value1" key2="value2">text</node>`,
		},
		{
			name: `escaping_attributes`,
			setup: func() XMLWriter {
				node := CreateElement("node")
				node.AddAttribute("", "k1", `val"ue`)
				node.AddAttribute("", "k2", "val'ue")
				return node
			},
			want: `<node k1="val\"ue" k2="val'ue"></node>`,
		},
		{
			name: `update_name_with_cdata_text`,
			setup: func() XMLWriter {
				node := CreateElement("node")
				node.SetName("new_name")
				node.SetText("text", true, NoEscaping)
				return node
			},
			want: `<new_name><![CDATA[text]]></new_name>`,
		},
		{
			name: `escaping_text`,
			setup: func() XMLWriter {
				node := CreateElement("node")
				node.SetText("<new & text>", true, XMLEscapeMode)
				return node
			},
			want: `<node><![CDATA[&lt;new &amp; text&gt;]]></node>`,
		},
		{
			name: `unescaping_text`,
			setup: func() XMLWriter {
				node := CreateElement("node")
				node.SetText("&lt;new &amp; text&gt;", true, XMLUnescapeMode)
				return node
			},
			want: `<node><![CDATA[<new & text>]]></node>`,
		},
		{
			name: `node_namespace`,
			setup: func() XMLWriter {
				node := CreateElement("node")
				node.SetNamespace("ns")
				node.SetText("new text", false, NoEscaping)
				return node
			},
			want: `<ns:node>new text</ns:node>`,
		},
		{
			name: `nested_node`,
			setup: func() XMLWriter {
				node := CreateElement("a").
					AddChild(CreateElement("b").
						AddChild(CreateElement("c").SetText("cdata", false, NoEscaping)).
						AddChild(CreateElement("d").SetText("ddata", false, NoEscaping)))
				return node
			},
			want: `<a><b><c>cdata</c><d>ddata</d></b></a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			node := tt.setup()
			node.Write(buf, &WriteSettings{})
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func Test_XMLTextElement(t *testing.T) {
	tests := []struct {
		name  string
		setup func() XMLWriter
		want  string
	}{
		{
			name: `simple_text`,
			setup: func() XMLWriter {
				return NewXMLText("text", false, NoEscaping)
			},
			want: `text`,
		},
		{
			name: `cdata_text`,
			setup: func() XMLWriter {
				return NewXMLText("text", true, NoEscaping)
			},
			want: `<![CDATA[text]]>`,
		},
		{
			name: `escaping_text`,
			setup: func() XMLWriter {
				return NewXMLText("<new & text>", false, XMLEscapeMode)
			},
			want: `&lt;new &amp; text&gt;`,
		},
		{
			name: `escaping_text`,
			setup: func() XMLWriter {
				return NewXMLText("&lt;new &amp; text&gt;", false, XMLUnescapeMode)
			},
			want: `<new & text>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			node := tt.setup()
			node.Write(buf, &WriteSettings{})
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func Test_XMLTextFunc(t *testing.T) {
	tests := []struct {
		name  string
		setup func() XMLWriter
		want  string
	}{
		{
			name: `nil function`,
			setup: func() XMLWriter {
				return NewXmlTextFunc(false, nil)
			},
			want: ``,
		},
		{
			name: `simple_text`,
			setup: func() XMLWriter {
				return NewXmlTextFunc(
					false,
					func(w Writer, _ *WriteSettings, args ...any) {
						w.WriteString(args[0].(string))
					},
					"text")
			},
			want: `text`,
		},
		{
			name: `multiple_args`,
			setup: func() XMLWriter {
				return NewXmlTextFunc(
					false,
					func(w Writer, ws *WriteSettings, args ...any) {
						w.WriteString(args[0].(string))
						w.WriteByte(':')
						args[1].(XMLWriter).Write(w, ws)
					},
					"text", NewXMLText("new_text", false, NoEscaping))
			},
			want: `text:new_text`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			node := tt.setup()
			node.Write(buf, &WriteSettings{})
			assert.Equal(t, tt.want, buf.String())
		})
	}
}
