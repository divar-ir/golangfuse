package golangfuse_test

import (
	"fmt"
	"testing"

	"github.com/divar-ir/golangfuse"
	"github.com/stretchr/testify/suite"
)

type UtilsTest struct {
	suite.Suite
}

func (s *UtilsTest) TestJinjaToGoTemplateShouldReturnVariablesWithGolangFormat() {
	// Given
	jinjaVarialbeVariants := []string{
		"{{myVar}}",
		"{{myVar }}",
		"{{ myVar}}",
		"{{ myVar }}",
	}

	for _, v := range jinjaVarialbeVariants {
		source := fmt.Sprintf("This is system prompt %s variable.", v)

		// When
		target := golangfuse.JinjaToGoTemplate(source)

		// Then
		s.Require().Equal("This is system prompt {{.myVar}} variable.", target)
	}
}

func TestUtils(t *testing.T) {
	suite.Run(t, new(UtilsTest))
}
