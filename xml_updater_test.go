package fastxml

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXMLUpdater(t *testing.T) {
	xmlDoc := []byte(`
<a>
    <b>b-data</b>
    <c>c-data</c>
    <d>
        <e>e-data</e>
		<c>c-data</c>
    </d>
	<f>
		<g gk1="true" gk2="deleteme" gk3="10">g-data</g>
	</f>
</a>`)

	reader := NewXMLReader(nil)
	if err := reader.Parse(xmlDoc); err != nil {
		t.Errorf("xml parsing error: %v", err.Error())
		return
	}

	elementB := reader.SelectElement(nil, "a", "b")
	elementF := reader.SelectElement(nil, "a", "f")
	elementG := reader.SelectElement(nil, "a", "f", "g")

	//xmlUpdater
	updater := NewXMLUpdater(reader, WriteSettings{})

	//remove elements
	updater.RemoveElement(reader.SelectElement(nil, "a", "c"))

	//replace full element
	updater.ReplaceElement(elementB, CreateElement("new_b").SetText("new_b_data", false, NoEscaping))

	//append or prepend new xml tag
	updater.PrependElement(elementF, CreateElement("f1").SetText("prepend_data", false, NoEscaping))
	updater.AppendElement(elementF, CreateElement("f2").SetText("append_data", false, NoEscaping))

	//append or prepend new xml tag in existing which has text
	updater.PrependElement(elementG, CreateElement("g1").SetText("prepend_tag", false, NoEscaping))
	updater.AppendElement(elementG, CreateElement("g2").SetText("append_tag", false, NoEscaping))

	//update text
	updater.UpdateText(elementG, "new-g-data", false, NoEscaping)

	//add new attribute
	updater.AddAttribute(elementF, "", "fk1", "fv1")

	//update attribute name and value
	gk1 := reader.SelectAttr(elementG, "gk1")
	//updater.UpdateAttributeName(gk1, "gk11")
	updater.UpdateAttributeValue(gk1, "false")

	//remove attribute
	gk2 := reader.SelectAttr(elementG, "gk2")
	updater.RemoveAttribute(gk2)

	//Build Updated XML File
	buf := bytes.Buffer{}
	updater.Build(&buf)

	t.Logf("\nOriginal XML:%s", xmlDoc)
	t.Logf("\n\nUpdated XML:%s", buf.String())
}

func TestXMLUpdater_AppendElement(t *testing.T) {
	type args struct {
		in         string
		operations func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil_element",
			args: args{
				in: ``,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AppendElement(nil, nil)
				},
			},
			want: ``,
		},
		{
			name: "nil_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AppendElement(reader.SelectElement(nil, "a"), nil)
				},
			},
			want: `<a></a>`,
		},
		{
			name: "append_inline_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AppendElement(reader.SelectElement(nil, "a"), CreateElement("").SetText("<empty_tag/>", false, NoEscaping))
				},
			},
			want: `<a><empty_tag/></a>`,
		},
		{
			name: "append_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AppendElement(reader.SelectElement(nil, "a"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a><tag>tagdata</tag></a>`,
		},
		{
			name: "empty_element_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AppendElement(nil, CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a></a>`,
		},
		{
			name: "tag_with_text",
			args: args{
				in: `<a>test_data</a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AppendElement(reader.SelectElement(nil, "a"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a>test_data<tag>tagdata</tag></a>`,
		},
		{
			name: "nested_tag_1",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AppendElement(reader.SelectElement(nil, "a", "b", "c"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a><b><c>cdata<tag>tagdata</tag></c></b></a>`,
		},
		{
			name: "nested_tag_2",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AppendElement(reader.SelectElement(nil, "a", "b"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a><b><c>cdata</c><tag>tagdata</tag></b></a>`,
		},
		{
			name: "multiple_elements",
			args: args{
				in: `<a><b>one</b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					elementA := reader.SelectElement(nil, "a")
					xu.AppendElement(elementA, CreateElement("b").SetText("two", false, NoEscaping))
					xu.AppendElement(elementA, CreateElement("b").SetText("three", false, NoEscaping))
				},
			},
			want: `<a><b>one</b><b>two</b><b>three</b></a>`,
		},
		{
			name: "multiple_nested_elements",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					elementB := reader.SelectElement(nil, "a", "b")
					elementC := reader.SelectElement(nil, "a", "b", "c")
					xu.AppendElement(elementB, CreateElement("b1").SetText("b1_data", false, NoEscaping))
					xu.AppendElement(elementC, CreateElement("c1").SetText("c1_data", false, NoEscaping))
				},
			},
			want: `<a><b><c>cdata<c1>c1_data</c1></c><b1>b1_data</b1></b></a>`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.operations(xu, reader)

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestXMLUpdater_PrependElement(t *testing.T) {
	type args struct {
		in         string
		operations func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil_element",
			args: args{
				in: ``,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.PrependElement(nil, nil)
				},
			},
			want: ``,
		},
		{
			name: "nil_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.PrependElement(reader.SelectElement(nil, "a"), nil)
				},
			},
			want: `<a></a>`,
		},
		/*
			{
				name: "prepend_inline_tag",
				args: args{
					in: `<a ak1="av1"/>`,
					operations: func(xu *XMLUpdater, reader *XMLReader) {
						reader := NewXMLReader(nil)
						_ = reader.Parse(in)
						xu.PrependElement(reader.FindElement(nil, "a"), `<tag>tagdata</tag>`)
					},
				},
				want: `<a ak1="av1"><tag>tagdata</tag></a>`,
			},
		*/
		{
			name: "prepend_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.PrependElement(reader.SelectElement(nil, "a"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a><tag>tagdata</tag></a>`,
		},
		{
			name: "empty_element_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.PrependElement(nil, CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a></a>`,
		},
		{
			name: "tag_with_text",
			args: args{
				in: `<a>test_data</a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.PrependElement(reader.SelectElement(nil, "a"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a><tag>tagdata</tag>test_data</a>`,
		},
		{
			name: "nested_tag_1",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.PrependElement(reader.SelectElement(nil, "a", "b", "c"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a><b><c><tag>tagdata</tag>cdata</c></b></a>`,
		},
		{
			name: "nested_tag_2",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.PrependElement(reader.SelectElement(nil, "a", "b"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a><b><tag>tagdata</tag><c>cdata</c></b></a>`,
		},
		{
			name: "multiple_elements",
			args: args{
				in: `<a><b>one</b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					elementA := reader.SelectElement(nil, "a")
					xu.PrependElement(elementA, CreateElement("b").SetText("two", false, NoEscaping))
					xu.PrependElement(elementA, CreateElement("b").SetText("three", false, NoEscaping))
				},
			},
			want: `<a><b>two</b><b>three</b><b>one</b></a>`,
		},
		{
			name: "multiple_nested_elements",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					elementB := reader.SelectElement(nil, "a", "b")
					elementC := reader.SelectElement(nil, "a", "b", "c")
					xu.PrependElement(elementB, CreateElement("b1").SetText("b1_data", false, NoEscaping))
					xu.PrependElement(elementC, CreateElement("c1").SetText("c1_data", false, NoEscaping))
				},
			},
			want: `<a><b><b1>b1_data</b1><c><c1>c1_data</c1>cdata</c></b></a>`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.operations(xu, reader)

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestXMLUpdater_ReplaceElement(t *testing.T) {
	type args struct {
		in         string
		operations func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "cannot_replace_in_empty_tag",
			args: args{
				in:         ``,
				operations: func(xu *XMLUpdater, reader *XMLReader) {},
			},
			want: ``,
		},
		{
			name: "replace_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.ReplaceElement(reader.SelectElement(nil, "a"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<tag>tagdata</tag>`,
		},
		{
			name: "replace_inline_tag",
			args: args{
				in: `<a/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.ReplaceElement(reader.SelectElement(nil, "a"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<tag>tagdata</tag>`,
		},
		{
			name: "empty_element_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.PrependElement(nil, CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a></a>`,
		},
		{
			name: "tag_with_text",
			args: args{
				in: `<a ak1="av1">test_data</a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.ReplaceElement(reader.SelectElement(nil, "a"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<tag>tagdata</tag>`,
		},
		{
			name: "nested_tag_1",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.ReplaceElement(reader.SelectElement(nil, "a", "b", "c"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a><b><tag>tagdata</tag></b></a>`,
		},
		{
			name: "nested_tag_2",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.ReplaceElement(reader.SelectElement(nil, "a", "b"), CreateElement("tag").SetText("tagdata", false, NoEscaping))
				},
			},
			want: `<a><tag>tagdata</tag></a>`,
		},
		/*
			{
				name: "invalid_replace_one_element_multiple_times",
				args: args{
					in: `<a><b>one</b></a>`,
					operations: func(xu *XMLUpdater, reader *XMLReader) {
						reader := NewXMLReader(nil)
						_ = reader.Parse(in)
						elementA := reader.FindElement(nil, "a")
						xu.ReplaceElement(elementA, NewXMLTag("b", "two"))
						xu.ReplaceElement(elementA, NewXMLTag("b", "three"))
					},
				},
				want: ``,
			},
			{
				name: "invalid_replace_overlapping_elements",
				args: args{
					in: `<a><b><c>cdata</c><d>ddata</d></b></a>`,
					operations: func(xu *XMLUpdater, reader *XMLReader) {
						reader := NewXMLReader(nil)
						_ = reader.Parse(in)
						elementB := reader.FindElement(nil, "a", "b")
						elementC := reader.FindElement(nil, "a", "b", "c")
						xu.ReplaceElement(elementB, NewXMLTag("b1", "b1_data"))
						xu.ReplaceElement(elementC, NewXMLTag("c1", "c1_data"))
					},
				},
				want: ``,
			},
		*/
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.operations(xu, reader)

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestXMLUpdater_RemoveElement(t *testing.T) {
	type args struct {
		in         string
		operations func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "cannot_remove_in_empty_tag",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveElement(nil)
				},
			},
			want: `<a></a>`,
		},
		{
			name: "remove_root_element",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveElement(reader.SelectElement(nil, "a"))
				},
			},
			want: ``,
		},
		{
			name: "remove_nested_element-1",
			args: args{
				in: `<a><b>bdata</b><c>cdata</c></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveElement(reader.SelectElement(nil, "a", "b"))
				},
			},
			want: `<a><c>cdata</c></a>`,
		},
		{
			name: "remove_nested_element-2",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveElement(reader.SelectElement(nil, "a", "b"))
				},
			},
			want: `<a></a>`,
		},
		{
			name: "remove_nested_element-3",
			args: args{
				in: `<a><b><c>cdata</c></b></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveElement(reader.SelectElement(nil, "a", "b", "c"))
				},
			},
			want: `<a><b></b></a>`,
		},
		{
			name: "remove_multiple_elements",
			args: args{
				in: `<a><b>bdata</b><c>cdata</c><d>ddata</d></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveElement(reader.SelectElement(nil, "a", "b"))
					xu.RemoveElement(reader.SelectElement(nil, "a", "d"))
				},
			},
			want: `<a><c>cdata</c></a>`,
		},
		{
			name: "remove_nested_operation",
			args: args{
				in: `<a><b><c>cdata</c></b><d>ddata</d></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveElement(reader.SelectElement(nil, "a", "b"))
					xu.RemoveElement(reader.SelectElement(nil, "a", "b", "c"))
				},
			},
			want: `<a><d>ddata</d></a>`,
		},
		{
			name: "remove_inline_element",
			args: args{
				in: `<a><c>cdata</c><b/><d>ddata</d></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveElement(reader.SelectElement(nil, "a", "b"))
				},
			},
			want: `<a><c>cdata</c><d>ddata</d></a>`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.operations(xu, reader)

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestXMLUpdater_UpdateText(t *testing.T) {
	type args struct {
		in         string
		operations func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil_element",
			args: args{
				in: ``,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.UpdateText(nil, "", false, NoEscaping)
				},
			},
			want: ``,
		},
		{
			name: "newtext",
			args: args{
				in: `<a>adata</a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.UpdateText(reader.SelectElement(nil, "a"), "newdata", false, NoEscaping)
				},
			},
			want: `<a>newdata</a>`,
		},
		{
			name: "newtext_cdata",
			args: args{
				in: `<a>adata</a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.UpdateText(reader.SelectElement(nil, "a"), "new data", true, NoEscaping)
				},
			},
			want: `<a><![CDATA[new data]]></a>`,
		},
		{
			name: "newtext_cdata",
			args: args{
				in: `<a>adata</a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.UpdateText(reader.SelectElement(nil, "a"), "<new & data>", true, XMLEscapeMode)
				},
			},
			want: `<a><![CDATA[&lt;new &amp; data&gt;]]></a>`,
		},
		{
			name: "newtext_cdata",
			args: args{
				in: `<a>adata</a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.UpdateText(reader.SelectElement(nil, "a"), "&lt;new &amp; data&gt;", true, XMLUnescapeMode)
				},
			},
			want: `<a><![CDATA[<new & data>]]></a>`,
		},
		{
			name: "inline_node",
			args: args{
				in: `<a/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.UpdateText(reader.SelectElement(nil, "a"), "new data", true, NoEscaping)
				},
			},
			want: `<a><![CDATA[new data]]></a>`,
		},
		{
			name: "inline_node_writeSettings.ExpandInline",
			args: args{
				in: `<a/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.writeSettings.ExpandInline = true
					xu.UpdateText(reader.SelectElement(nil, "a"), "new data", true, NoEscaping)
				},
			},
			want: `<a/>`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.operations(xu, reader)

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestXMLUpdater_AddAttribute(t *testing.T) {
	type args struct {
		in         string
		operations func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil_element",
			args: args{
				in: ``,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AddAttribute(nil, "ns", "key", "value")
				},
			},
			want: ``,
		},
		{
			name: "add_attribute",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AddAttribute(reader.SelectElement(nil, "a"), "ns", "key", "value")
				},
			},
			want: `<a ns:key="value"></a>`,
		},
		{
			name: "without_namespace",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AddAttribute(reader.SelectElement(nil, "a"), "", "key", "value")
				},
			},
			want: `<a key="value"></a>`,
		},
		{
			name: "prepend_new_attribute",
			args: args{
				in: `<a key="value"></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AddAttribute(reader.SelectElement(nil, "a"), "", "key1", "value1")
				},
			},
			want: `<a key1="value1" key="value"></a>`,
		},
		{
			name: "multiple_attribute",
			args: args{
				in: `<a></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AddAttribute(reader.SelectElement(nil, "a"), "", "key", "value")
					xu.AddAttribute(reader.SelectElement(nil, "a"), "", "key1", "value1")
				},
			},
			want: `<a key="value" key1="value1"></a>`,
		},
		{
			name: "inline_element",
			args: args{
				in: `<a/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AddAttribute(reader.SelectElement(nil, "a"), "", "key", "value")
				},
			},
			want: `<a key="value"/>`,
		},
		{
			name: "escaping_double_quote",
			args: args{
				in: `<a/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AddAttribute(reader.SelectElement(nil, "a"), "", "key", `val"ue`)
				},
			},
			want: `<a key="val\"ue"/>`,
		},
		{
			name: "escaping_single_quote",
			args: args{
				in: `<a/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.AddAttribute(reader.SelectElement(nil, "a"), "", "key", `val'ue`)
				},
			},
			want: `<a key="val'ue"/>`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.operations(xu, reader)

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestXMLUpdater_RemoveAttribute(t *testing.T) {
	type args struct {
		in         string
		operations func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil_attribute",
			args: args{
				in: ``,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveAttribute(nil)
				},
			},
			want: ``,
		},
		{
			name: "remove_first_attribute",
			args: args{
				in: `<a key1="value1"></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.RemoveAttribute(reader.SelectAttr(reader.SelectElement(nil, "a"), "key1"))
				},
			},
			want: `<a></a>`,
		},
		{
			name: "remove_multiple_attribute",
			args: args{
				in: `<a key1="value1" key2="value2" key3="value3"></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					aElement := reader.SelectElement(nil, "a")
					xu.RemoveAttribute(reader.SelectAttr(aElement, "key1"))
					xu.RemoveAttribute(reader.SelectAttr(aElement, "key3"))
				},
			},
			want: `<a key2="value2"></a>`,
		},
		{
			name: "inline_element",
			args: args{
				in: `<a key1="value1" key2="value2" key3="value3"/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					aElement := reader.SelectElement(nil, "a")
					xu.RemoveAttribute(reader.SelectAttr(aElement, "key1"))
					xu.RemoveAttribute(reader.SelectAttr(aElement, "key3"))
				},
			},
			want: `<a key2="value2"/>`,
		},
		{
			name: "inline_element_remove_all",
			args: args{
				in: `<a key1="value1"/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					aElement := reader.SelectElement(nil, "a")
					xu.RemoveAttribute(reader.SelectAttr(aElement, "key1"))
				},
			},
			want: `<a/>`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.operations(xu, reader)

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestXMLUpdater_UpdateAttributeValue(t *testing.T) {
	type args struct {
		in         string
		operations func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil_element",
			args: args{
				in: ``,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.UpdateAttributeValue(nil, "")
				},
			},
			want: ``,
		},
		{
			name: "empty_value",
			args: args{
				in: `<a key="value"></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					aElement := reader.SelectElement(nil, "a")
					xu.UpdateAttributeValue(reader.SelectAttr(aElement, "key"), "")
				},
			},
			want: `<a key=""></a>`,
		},
		{
			name: "empty_value",
			args: args{
				in: `<a key="value"></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					aElement := reader.SelectElement(nil, "a")
					xu.UpdateAttributeValue(reader.SelectAttr(aElement, "key"), "")
				},
			},
			want: `<a key=""></a>`,
		},
		{
			name: "replace_value",
			args: args{
				in: `<a key="value"></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					aElement := reader.SelectElement(nil, "a")
					xu.UpdateAttributeValue(reader.SelectAttr(aElement, "key"), "new_value")
				},
			},
			want: `<a key="new_value"></a>`,
		},
		{
			name: "inline_element",
			args: args{
				in: `<a key="value"/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					aElement := reader.SelectElement(nil, "a")
					xu.UpdateAttributeValue(reader.SelectAttr(aElement, "key"), "new_value")
				},
			},
			want: `<a key="new_value"/>`,
		},
		{
			name: "escaping_value",
			args: args{
				in: `<a key="value"></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					aElement := reader.SelectElement(nil, "a")
					xu.UpdateAttributeValue(reader.SelectAttr(aElement, "key"), `new_"value`)
				},
			},
			want: `<a key="new_\"value"></a>`,
		},
		{
			name: "single_quote_escaping_value",
			args: args{
				in: `<a key="value"></a>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					aElement := reader.SelectElement(nil, "a")
					xu.UpdateAttributeValue(reader.SelectAttr(aElement, "key"), `new_'value`)
				},
			},
			want: `<a key="new_'value"></a>`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.operations(xu, reader)

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestXMLUpdater_expandInline(t *testing.T) {
	type args struct {
		in         string
		operations func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil_element",
			args: args{
				in: ``,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.expandInline(nil)
				},
			},
			want: ``,
		},
		{
			name: "expand_inline",
			args: args{
				in: `<a/>`,
				operations: func(xu *XMLUpdater, reader *XMLReader) {
					xu.expandInline(reader.SelectElement(nil, "a"))
				},
			},
			want: `<a></a>`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.operations(xu, reader)

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestXMLUpdater_ApplyXMLSettingsOperations(t *testing.T) {
	type args struct {
		in    string
		setup func(xu *XMLUpdater, reader *XMLReader)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no_settings_enabled",
			args: args{
				in: `<a><b/><c/><d>ddata</d><e>edata</e><f><![CDATA[fdata]]></f><g><![CDATA[gdata]]></g></a>`,
				setup: func(xu *XMLUpdater, reader *XMLReader) {
					xu.writeSettings.CDATAWrap = false
					xu.writeSettings.ExpandInline = false
				},
			},
			want: `<a><b/><c/><d>ddata</d><e>edata</e><f><![CDATA[fdata]]></f><g><![CDATA[gdata]]></g></a>`,
		},
		{
			name: "cdata_enabled",
			args: args{
				in: `<a><b/><c/><d>ddata</d><e>edata</e><f><![CDATA[fdata]]></f><g><![CDATA[gdata]]></g></a>`,
				setup: func(xu *XMLUpdater, reader *XMLReader) {
					xu.writeSettings.CDATAWrap = true
					xu.writeSettings.ExpandInline = false
				},
			},
			want: `<a><b/><c/><d><![CDATA[ddata]]></d><e><![CDATA[edata]]></e><f><![CDATA[fdata]]></f><g><![CDATA[gdata]]></g></a>`,
		},
		{
			name: "inline_enabled",
			args: args{
				in: `<a><b/><c/><d>ddata</d><e>edata</e><f><![CDATA[fdata]]></f><g><![CDATA[gdata]]></g></a>`,
				setup: func(xu *XMLUpdater, reader *XMLReader) {
					xu.writeSettings.CDATAWrap = false
					xu.writeSettings.ExpandInline = true
				},
			},
			want: `<a><b></b><c></c><d>ddata</d><e>edata</e><f><![CDATA[fdata]]></f><g><![CDATA[gdata]]></g></a>`,
		},
		{
			name: "cdata_inline_enabled",
			args: args{
				in: `<a><b/><c/><d>ddata</d><e>edata</e><f><![CDATA[fdata]]></f><g><![CDATA[gdata]]></g></a>`,
				setup: func(xu *XMLUpdater, reader *XMLReader) {
					xu.writeSettings.CDATAWrap = true
					xu.writeSettings.ExpandInline = true
				},
			},
			want: `<a><b></b><c></c><d><![CDATA[ddata]]></d><e><![CDATA[edata]]></e><f><![CDATA[fdata]]></f><g><![CDATA[gdata]]></g></a>`,
		},
		{
			name: "cdata_unescapemode",
			args: args{
				in: `<a>&lt;new &amp; data&gt;</a>`,
				setup: func(xu *XMLUpdater, reader *XMLReader) {
					xu.writeSettings.CDATAWrap = true
					xu.writeSettings.ExpandInline = true
				},
			},
			want: `<a><![CDATA[<new & data>]]></a>`,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewXMLReader(nil)
			_ = reader.Parse([]byte(tt.args.in))

			xu := NewXMLUpdater(reader, WriteSettings{})
			tt.args.setup(xu, reader)

			xu.applyXMLSettings()

			//rebuild buffer
			out := bytes.Buffer{}
			xu.Build(&out)
			assert.Equal(t, tt.want, out.String())
		})
	}
}
