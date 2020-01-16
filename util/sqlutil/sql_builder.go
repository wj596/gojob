/*
 * Copyright 2020-2021 the original author(https://github.com/wj596)
 *
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * </p>
 */
package sqlutil

import (
	"bytes"
	"fmt"
	"strconv"
)

const (
	statementTypeDelete = "delete"
	statementTypeInsert = "insert"
	statementTypeSelect = "select"
	statementTypeUpdate = "update"
	and                 = ") \nAND ("
	or                  = ") \nOR ("
)

type SqlBuilder struct {
	statementType  string
	distinct       bool
	sets           []string
	selects        []string
	tables         []string
	join           []string
	innerJoin      []string
	outerJoin      []string
	leftOuterJoin  []string
	rightOuterJoin []string
	where          []string
	having         []string
	groupBy        []string
	orderBy        []string
	lastList       int // 1：where 2:having
	columns        []string
	values         []string
	limit          int
	offset         int
}

func NewSqlBuilder() *SqlBuilder {
	return new(SqlBuilder)
}

func (this *SqlBuilder) SELECT(columns string) *SqlBuilder {
	this.statementType = statementTypeSelect
	this.selects = append(this.selects, columns)
	return this
}

func (this *SqlBuilder) SELECT_DISTINCT(columns string) *SqlBuilder {
	this.distinct = true
	this.statementType = statementTypeSelect
	this.selects = append(this.selects, columns)
	return this
}

func (this *SqlBuilder) REST_SELECT() *SqlBuilder {
	this.selects = make([]string, 0)
	return this
}

func (this *SqlBuilder) UPDATE(table string) *SqlBuilder {
	this.statementType = statementTypeUpdate
	this.tables = append(this.tables, table)
	return this
}

func (this *SqlBuilder) SET(vals string) *SqlBuilder {
	this.sets = append(this.sets, vals)
	return this
}

func (this *SqlBuilder) INSERT_INTO(tableName string) *SqlBuilder {
	this.statementType = statementTypeInsert
	this.tables = append(this.tables, tableName)
	return this
}

func (this *SqlBuilder) VALUES(cols string, vals string) *SqlBuilder {
	this.columns = append(this.columns, cols)
	this.values = append(this.values, vals)
	return this
}

func (this *SqlBuilder) DELETE_FROM(tableName string) *SqlBuilder {
	this.statementType = statementTypeDelete
	this.tables = append(this.tables, tableName)
	return this
}

func (this *SqlBuilder) FROM(tableName string) *SqlBuilder {
	this.tables = append(this.tables, tableName)
	return this
}

func (this *SqlBuilder) JOIN(table string) *SqlBuilder {
	this.join = append(this.join, table)
	return this
}

func (this *SqlBuilder) INNER_JOIN(table string) *SqlBuilder {
	this.innerJoin = append(this.innerJoin, table)
	return this
}

func (this *SqlBuilder) LEFT_OUTER_JOIN(table string) *SqlBuilder {
	this.leftOuterJoin = append(this.leftOuterJoin, table)
	return this
}

func (this *SqlBuilder) RIGHT_OUTER_JOIN(table string) *SqlBuilder {
	this.rightOuterJoin = append(this.rightOuterJoin, table)
	return this
}

func (this *SqlBuilder) OUTER_JOIN(table string) *SqlBuilder {
	this.outerJoin = append(this.outerJoin, table)
	return this
}

func (this *SqlBuilder) WHERE(conditions string) *SqlBuilder {
	this.where = append(this.where, conditions)
	this.lastList = 1
	return this
}

func (this *SqlBuilder) WHEREF(tpl string, args ...interface{}) *SqlBuilder {
	this.where = append(this.where, fmt.Sprintf(tpl, args...))
	this.lastList = 1
	return this
}

func (this *SqlBuilder) WHEREF_NECESSARY(need bool, tpl string, args ...interface{}) *SqlBuilder {
	if need {
		this.WHEREF(tpl, args...)
	}
	return this
}

func (this *SqlBuilder) OR() *SqlBuilder {
	if this.lastList == 1 {
		this.WHERE(or)
	}
	if this.lastList == 2 {
		this.HAVING(or)
	}
	return this
}

func (this *SqlBuilder) AND() *SqlBuilder {
	if this.lastList == 1 {
		this.WHERE(and)
	}
	if this.lastList == 2 {
		this.HAVING(and)
	}
	return this
}

func (this *SqlBuilder) GROUP_BY(cols string) *SqlBuilder {
	this.groupBy = append(this.groupBy, cols)
	return this
}

func (this *SqlBuilder) HAVING(conditions string) *SqlBuilder {
	this.having = append(this.having, conditions)
	this.lastList = 2
	return this
}

func (this *SqlBuilder) ORDER_BY(cols string) *SqlBuilder {
	this.orderBy = append(this.orderBy, cols)
	return this
}

func (this *SqlBuilder) LIMIT(l int, o int) *SqlBuilder {
	this.limit = l
	this.offset = o
	return this
}

func (this *SqlBuilder) Sql() string {
	buffer := new(bytes.Buffer)
	switch this.statementType {
	case statementTypeSelect:
		this.selectSql(buffer)
	case statementTypeUpdate:
		this.updateSql(buffer)
	case statementTypeInsert:
		this.insertSql(buffer)
	case statementTypeDelete:
		this.deleteSql(buffer)
	}
	if this.limit > 0 {
		buffer.WriteString("\n LIMIT ")
		buffer.WriteString(strconv.Itoa(this.offset))
		buffer.WriteString(",")
		buffer.WriteString(strconv.Itoa(this.limit))
	}
	return buffer.String()
}

func (this *SqlBuilder) selectSql(builder *bytes.Buffer) {
	if this.distinct {
		this.clause(builder, "SELECT DISTINCT", this.selects, "", "", ", ")
	} else {
		this.clause(builder, "SELECT", this.selects, "", "", ", ")
	}
	this.clause(builder, "FROM", this.tables, "", "", ", ")

	this.clause(builder, "JOIN", this.join, "", "", "\nJOIN ")
	this.clause(builder, "INNER JOIN", this.innerJoin, "", "", "\nINNER JOIN ")
	this.clause(builder, "OUTER JOIN", this.outerJoin, "", "", "\nOUTER JOIN ")
	this.clause(builder, "LEFT OUTER JOIN", this.leftOuterJoin, "", "", "\nLEFT OUTER JOIN ")
	this.clause(builder, "RIGHT OUTER JOIN", this.rightOuterJoin, "", "", "\nRIGHT OUTER JOIN ")

	this.clause(builder, "WHERE", this.where, "(", ")", " AND ")
	this.clause(builder, "GROUP BY", this.groupBy, "", "", ", ")
	this.clause(builder, "HAVING", this.having, "(", ")", " AND ")
	this.clause(builder, "ORDER BY", this.orderBy, "", "", ", ")
}

func (this *SqlBuilder) updateSql(builder *bytes.Buffer) {
	this.clause(builder, "UPDATE", this.tables, "", "", "")
	this.clause(builder, "SET", this.sets, "", "", ", ")
	this.clause(builder, "WHERE", this.where, "(", ")", " AND ")
}

func (this *SqlBuilder) insertSql(builder *bytes.Buffer) {
	this.clause(builder, "INSERT INTO", this.tables, "", "", "")
	this.clause(builder, "", this.columns, "(", ")", ", ")
	this.clause(builder, "VALUES", this.values, "(", ")", ", ")
}

func (this *SqlBuilder) deleteSql(builder *bytes.Buffer) {
	this.clause(builder, "DELETE FROM", this.tables, "", "", "")
	this.clause(builder, "WHERE", this.where, "(", ")", " AND ")
}

func (this *SqlBuilder) clause(builder *bytes.Buffer,
	keyword string, parts []string, open string, close string, conjunction string) {

	if len(parts) > 0 {
		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(keyword)
		builder.WriteString(" ")
		builder.WriteString(open)
		last := "________"
		for i, part := range parts {
			if i > 0 && part != and && part != or && last != and && last != or {
				builder.WriteString(conjunction)
			}
			builder.WriteString(part)
			last = part
		}
		builder.WriteString(close)
	}
}
