package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/getwtxt/registry"
)

func Test_parseQueryOut(t *testing.T) {
	initTestConf()

	urls := "https://gbmor.dev/twtxt.txt"
	nick := "gbmor"

	out, _, err := registry.GetTwtxt(urls)
	if err != nil {
		t.Errorf("Couldn't set up test: %v\n", err)
	}

	statusmap, err := registry.ParseUserTwtxt(out, nick, urls)
	if err != nil {
		t.Errorf("Couldn't set up test: %v\n", err)
	}

	twtxtCache.AddUser(nick, urls, "", net.ParseIP("127.0.0.1"), statusmap)

	t.Run("Parsing Status Query", func(t *testing.T) {
		data, err := twtxtCache.QueryAllStatuses()
		if err != nil {
			t.Errorf("%v\n", err)
		}

		out := parseQueryOut(data)

		conv := strings.Split(string(out), "\n")

		if !reflect.DeepEqual(data, conv) {
			t.Errorf("Pre- and Post- parseQueryOut data are inequal:\n%#v\n%#v\n", data, conv)
		}
	})
}

func Benchmark_parseQueryOut(b *testing.B) {
	initTestConf()

	urls := "https://gbmor.dev/twtxt.txt"
	nick := "gbmor"

	out, _, err := registry.GetTwtxt(urls)
	if err != nil {
		b.Errorf("Couldn't set up test: %v\n", err)
	}

	statusmap, err := registry.ParseUserTwtxt(out, nick, urls)
	if err != nil {
		b.Errorf("Couldn't set up test: %v\n", err)
	}

	twtxtCache.AddUser(nick, urls, "", net.ParseIP("127.0.0.1"), statusmap)

	data, err := twtxtCache.QueryAllStatuses()
	if err != nil {
		b.Errorf("%v\n", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parseQueryOut(data)
	}

}
