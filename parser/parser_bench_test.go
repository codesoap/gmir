package parser_test

import (
	"strings"
	"testing"

	"github.com/codesoap/gmir/parser"
)

const testGMI = `# Some blog
Welcome to my blog. It is filled with interesting doloribus ducimus quisquam dolor expedita possimus molestiae. Quis in ut nam nisi. Sunt esse dolorum quod vel corrupti nihil mollitia.

Hic exercitationem optio molestiae. Ipsam sequi animi eius voluptas doloribus ea nihil est. Debitis est a magni sit. Exercitationem voluptatem rerum illum dolores commodi neque nemo quae.

Nulla sit dolorem et minus sunt quo. Nulla laboriosam maiores ut rerum fugit minus consectetur voluptas. Eum aut natus dignissimos et vitae amet. In sint ullam ut minima velit illo ratione. Sunt quia sunt corporis iure. Aperiam exercitationem nulla quos.

Aspernatur eveniet voluptate quisquam sapiente deleniti cumque et. Quasi et iste quas cum suscipit et debitis rem. Error aut quos aliquam. Earum ut eveniet et dolores hic iure. Aut qui culpa soluta. Cum similique consectetur odit fuga qui accusantium explicabo.

It's got lists like:
* one.
* two.
* three (this is a somewhat longer list item to force wrapping of the line).

## Quotes
> Someone smart said something smart.

=> /ref1.gmi My first reference
=> /ref2.gmi My second reference

` + "```\nASCII ART\n```\n"

func BenchmarkParsing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := parser.Parse(strings.NewReader(testGMI)); err != nil {
			b.Errorf("Could not parse input: %v", err)
			break
		}
	}
}

func BenchmarkWrapping(b *testing.B) {
	lines, err := parser.Parse(strings.NewReader(testGMI))
	if err != nil {
		b.Errorf("Could not parse input: %v", err)
		return
	}
	for i := 0; i < b.N; i++ {
		for _, line := range lines {
			if wrappable, ok := line.(parser.WrappableLine); ok {
				wrappable.WrapIndexes(16)
			}
		}
	}
}
