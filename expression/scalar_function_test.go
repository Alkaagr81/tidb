// Copyright 2017 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package expression

import (
	"time"

	. "github.com/pingcap/check"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/mysql"
	"github.com/pingcap/tidb/sessionctx/stmtctx"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/util/chunk"
)

func (s *testEvaluatorSuite) TestScalarFunction(c *C) {
	a := &Column{
		UniqueID: 1,
		RetType:  types.NewFieldType(mysql.TypeDouble),
	}
	sc := &stmtctx.StatementContext{TimeZone: time.Local}
	sf := newFunction(ast.LT, a, NewOne())
	res, err := sf.MarshalJSON()
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, []byte{0x22, 0x6c, 0x74, 0x28, 0x43, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x23, 0x31, 0x2c, 0x20, 0x31, 0x29, 0x22})
	c.Assert(sf.IsCorrelated(), IsFalse)
	c.Assert(sf.ConstItem(s.ctx.GetSessionVars().StmtCtx), IsFalse)
	c.Assert(sf.Decorrelate(nil).Equal(s.ctx, sf), IsTrue)
	c.Assert(sf.HashCode(sc), DeepEquals, []byte{0x3, 0x4, 0x6c, 0x74, 0x1, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x5, 0xbf, 0xf0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})

	sf = NewValuesFunc(s.ctx, 0, types.NewFieldType(mysql.TypeLonglong))
	newSf, ok := sf.Clone().(*ScalarFunction)
	c.Assert(ok, IsTrue)
	c.Assert(newSf.FuncName.O, Equals, "values")
	c.Assert(newSf.RetType.Tp, Equals, mysql.TypeLonglong)
	_, ok = newSf.Function.(*builtinValuesIntSig)
	c.Assert(ok, IsTrue)
}

func (s *testEvaluatorSuite) TestIssue23309(c *C) {
	a := &Column{
		UniqueID: 1,
		RetType:  types.NewFieldType(mysql.TypeDouble),
	}
	a.RetType.Flag |= mysql.NotNullFlag
	null := NewNull()
	null.RetType = types.NewFieldType(mysql.TypeNull)
	sf, _ := newFunction(ast.NE, a, null).(*ScalarFunction)
	v, err := sf.GetArgs()[1].Eval(chunk.Row{})
	c.Assert(err, IsNil)
	c.Assert(v.IsNull(), IsTrue)
	c.Assert(mysql.HasNotNullFlag(sf.GetArgs()[1].GetType().Flag), IsFalse)
}

func (s *testEvaluatorSuite) TestScalarFuncs2Exprs(c *C) {
	a := &Column{
		UniqueID: 1,
		RetType:  types.NewFieldType(mysql.TypeDouble),
	}
	sf0, _ := newFunction(ast.LT, a, NewZero()).(*ScalarFunction)
	sf1, _ := newFunction(ast.LT, a, NewOne()).(*ScalarFunction)

	funcs := []*ScalarFunction{sf0, sf1}
	exprs := ScalarFuncs2Exprs(funcs)
	for i := range exprs {
		c.Assert(exprs[i].Equal(s.ctx, funcs[i]), IsTrue)
	}
}
