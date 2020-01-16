package sqlutil

import (
	"fmt"
	"testing"
)

func TestBuilderSelect(t *testing.T) {
	sb := NewSqlBuilder().SELECT("*").FROM("t_user").
		WHERE("name = 2").WHERE("age=12")
	fmt.Println(sb.Sql())
}

func TestBuilderOR(t *testing.T) {
	sb := NewSqlBuilder().SELECT("*").FROM("t_user").
		WHERE("name = 2").OR().WHERE("age=12")
	fmt.Println(sb.Sql())
}

func TestBuilderWhereF(t *testing.T) {
	sb := NewSqlBuilder().SELECT("*").FROM("t_user").
		WHEREF("name = '%s' and age = %d", "%wj%", 11)
	fmt.Println(sb.Sql())
}
