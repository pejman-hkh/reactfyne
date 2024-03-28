package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/pejman-hkh/gdp/gdp"
)

func convertToGoxFunc(tag *gdp.Tag, child string) string {
	attrs := ""
	pre := ""
	for key, value := range tag.Attrs() {
		r := regexp.MustCompile(`{{([^{}]*)}}`)
		matches := r.FindAllStringSubmatch(*value, -1)

		if len(matches) > 0 && len(matches[0]) > 1 {
			attrs += fmt.Sprintf("%s`%s` :%s", pre, key, matches[0][1])
		} else {
			attrs += fmt.Sprintf("%s`%s` :`%s`", pre, key, *value)
		}
		pre = ","
	}

	return fmt.Sprintf(`react.Run("%s", map[string]string{`+attrs+`}, %s)`, tag.TagName(), child)
}

func ToGo(tag *gdp.Tag) string {
	ret := ""
	pre := ""
	tag.Children().Each(func(i int, t *gdp.Tag) {
		if t.TagName() == "empty" {
			if t.Parent().TagName() == "document" {
				ret += t.Content()
			} else {
				content := t.Content()

				r := regexp.MustCompile(`(.*?){([^{}]*)}(.*?)`)
				matches := r.FindAllStringSubmatch(content, -1)
				if len(matches) > 0 {
					ra := ""
					for _, v := range matches {
						ra += "`" + v[1] + "`"
						if v[2] != "" {
							ra += "," + v[2]
						}
						if v[3] != "" {
							ra += "," + v[3]
						}
					}
					if ra != "" {
						ret += pre + ra
					}
				} else {
					ret += pre + "`" + content + "`"
				}
			}
		} else {
			childs := `[]any{`

			if t.Children().Length() > 0 {
				childs += ToGo(t)
			}
			childs += `}`
			ret += pre + convertToGoxFunc(t, childs)
		}
		if t.Parent().TagName() != "document" {
			pre = ", "
		}
	})
	return ret
}

func main() {
	dir := os.Args[1]

	files, _ := ioutil.ReadDir(dir)

	for _, file := range files {
		name := file.Name()
		len := len(name)

		if name[(len-3):len] == "gox" {
			fmt.Printf("%s\n", name)
			goxFile, _ := os.ReadFile(dir + name)
			document := gdp.Default(string(goxFile))

			out := ToGo(&document)

			ioutil.WriteFile(dir+name[0:len-1], []byte(out[:]), 0644)
		}

	}
}
