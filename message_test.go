package irc

import (
	"bytes"
	"strings"
	"testing"
	"encoding/json"
)

type mtest struct {
	line        []byte
	msg         Message
	shouldFail  bool
	skipMarshal bool
}

var (
	mtests = []mtest{

		mtest{
			[]byte(":nick!user@host.com PRIVMSG #42o3L1t3 :l0l sw4g 0m9 s0 c00l 42o"),
			Message{"nick!user@host.com", "PRIVMSG", []string{"#42o3L1t3"}, "l0l sw4g 0m9 s0 c00l 42o"},
			false,
			false,
		},

		mtest{
			[]byte("COMMAND :no arguments here!"),
			Message{"", "COMMAND", nil, "no arguments here!"},
			false,
			false,
		},

		mtest{
			[]byte("COMMAND nothing but arguments!"),
			Message{"", "COMMAND", []string{"nothing", "but", "arguments!"}, ""},
			false,
			false,
		},

		// empty - valid per rfc
		mtest{
			nil,
			Message{"", "", nil, ""},
			true, // must fail- empty lines are valid but the user needs to know
			false,
		},

		// empty prefix - short message
		mtest{
			[]byte(":"),
			Message{"", "", nil, ""},
			true,
			true,
		},

		// prefix, no command - short message
		mtest{
			[]byte(":foo"),
			Message{"foo", "", nil, ""},
			true,
			false,
		},

		/* no tags for now */
		/*
		   mtest{
		     []byte("@foo"),
		     Message{"@foo", "", []string{""}},
		     false,
		   },

		   mtest{
		     []byte("@foo :bar"),
		     Message{"", "", []string{""}},
		     false,
		   },
		*/

		mtest{
			[]byte("foo bar baz asdf"),
			Message{"", "foo", []string{"bar", "baz", "asdf"}, ""},
			false,
			false,
		},

		mtest{
			[]byte("foo bar baz :asdf quux"),
			Message{"", "foo", []string{"bar", "baz"}, "asdf quux"},
			false,
			false,
		},

		mtest{
			[]byte("foo bar baz"),
			Message{"", "foo", []string{"bar", "baz"}, ""},
			false,
			false,
		},

		mtest{
			[]byte("foo bar baz ::asdf"),
			Message{"", "foo", []string{"bar", "baz"}, ":asdf"},
			false,
			false,
		},

		mtest{
			[]byte(":test foo bar baz asdf"),
			Message{"test", "foo", []string{"bar", "baz", "asdf"}, ""},
			false,
			false,
		},

		mtest{
			[]byte(":test foo bar baz :asdf quux"),
			Message{"test", "foo", []string{"bar", "baz"}, "asdf quux"},
			false,
			false,
		},

		mtest{
			[]byte(":test foo bar baz"),
			Message{"test", "foo", []string{"bar", "baz"}, ""},
			false,
			false,
		},

		mtest{
			[]byte(":test foo bar baz ::asdf"),
			Message{"test", "foo", []string{"bar", "baz"}, ":asdf"},
			false,
			false,
		},

		mtest{
			[]byte(":foo bar"),
			Message{"foo", "bar", nil, ""},
			false,
			false,
		},

		mtest{
			[]byte(":foo :bar baz"),
			Message{"foo", "", nil, "bar baz"},
			true,
			false,
		},
	}
)

func TestUnmarshalText(t *testing.T) {
	for _, ut := range mtests {
		var m Message

		if err := m.Unmarshal(ut.line); err != nil {
			if ut.shouldFail {
				t.Logf("expected failure: on %q: %s", ut.line, err)
				continue
			}
			t.Fatalf("unmarshal failed on %q: %s", ut.line, err)
		}

		t.Logf("unmarshal(%q)\n\tgot  %#v\n\twant %#v", ut.line, m, ut.msg)

		if m.Prefix != ut.msg.Prefix {
			t.Fatalf("Bad prefix")
		}

		if strings.ToUpper(m.Command) != strings.ToUpper(ut.msg.Command) {
			t.Fatalf("Bad command")
		}

		if len(m.Parameters) != len(ut.msg.Parameters) {
			t.Fatalf("Bad Arguments count")
		}

		for i, arg := range m.Parameters {

			if arg != ut.msg.Parameters[i] {
				t.Fatalf("Bad argument %v", i)
			}
		}

		if m.Trailing != ut.msg.Trailing {
			t.Fatalf("Bad trailing")
		}
	}
}
func TestMarshalText(t *testing.T) {
	for _, mt := range mtests {
		if mt.skipMarshal {
			continue
		}

		line, err := mt.msg.Marshal()

		if err != nil && mt.shouldFail {
			t.Logf("expected failure on %s: %s", mt.msg, err)
			continue
		}

		if err != nil {
			t.Errorf("marshal %s failed: %s", mt.msg, err)
			continue
		}

		if bytes.Compare(line, mt.line) != 0 {
			t.Errorf("different lines:\n\tgot  %q\n\twant %q", line, mt.line)
		}
	}
}

var unmarshaltext_bench = []byte(":server.kevlar.net NOTICE user :*** This is a test")

func BenchmarkUnmarshalText(b *testing.B) {
	var m Message
	for i := 0; i < b.N; i++ {
		m.Unmarshal(unmarshaltext_bench)
	}

}

var unmarshaljson_bench = []byte(
	`{"Prefix":"server.kevlar.net","Command":"Notice","Parameters":["user"],"Trailing":"*** This is a test"}`)

func BenchmarkUnmarshalJson (b *testing.B) {

	for i := 0; i < b.N; i++ {
		var m Message
		json.Unmarshal(unmarshaljson_bench, &m)
	}
}

var marshaltext_bench = Message{"someguy!user@foo.bar.com", "PRIVMSG", []string{"#testing"}, "foo bar baz quux"}

func BenchmarkMarshalText(b *testing.B) {
	for i := 0; i < b.N; i++ {
		marshaltext_bench.Marshal()
	}
}
