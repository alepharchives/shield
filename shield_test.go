package shield

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func readDataSet(dataFile, labelFile string, t *testing.T) []string {
	d, err := ioutil.ReadFile("testdata/" + dataFile)
	if err != nil {
		t.Fatal(err)
	}
	l, err := ioutil.ReadFile("testdata/" + labelFile)
	if err != nil {
		t.Fatal(err)
	}
	dl := strings.Split(string(d), "\n")
	ll := strings.Split(string(l), "\n")
	x, y := len(dl), len(ll)
	if x != y {
		t.Fatal(x, y)
	}
	var a []string
	for i, v := range ll {
		k := strings.TrimSpace(v)
		if k != "" {
			a = append(a, fmt.Sprintf("%s %s", k, strings.TrimSpace(dl[i])))
		}
	}
	return a
}

func TestLearn(t *testing.T) {
	testData := readDataSet("testdata.txt", "testlabels.txt", t)
	trainData := readDataSet("traindata.txt", "trainlabels.txt", t)

	logger := log.New(os.Stderr, "", log.LstdFlags)
	store := NewRedisStore("127.0.0.1:6379", "", logger, "redis")
	tokenizer := NewEnglishTokenizer()

	sh := New(tokenizer, store)
	err := sh.Reset()
	if err != nil {
		t.Fatal(err)
	}
	// defer sh.Reset()

	// Run on test sets
	sets := []Set{}
	for _, v := range trainData {
		c := strings.SplitN(v, " ", 2)
		sets = append(sets, Set{
			Class: c[0],
			Text:  c[1],
		})
	}

	sh.BulkLearn(sets)

	var hit, miss int
	for _, v2 := range testData {
		c := strings.SplitN(v2, " ", 2)
		k, v := c[0], c[1]
		clz, err := sh.Classify(v)
		if err != nil {
			t.Fatal(err, "key:", k, "value:", v)
		}
		if clz != k {
			miss++
		} else {
			hit++
		}
	}

	// Test hit/miss ratio
	// TODO: Tweak this, where possible
	minHitRatio := 0.73
	hitRatio := (float64(hit) / float64(hit+miss))
	if hitRatio < minHitRatio {
		t.Fatalf("%d hits, %d misses (expected ratio %.2f, is %.2f)", hit, miss, minHitRatio, hitRatio)
	}
}
