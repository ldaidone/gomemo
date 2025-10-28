package memo

//
//import (
//	"context"
//	"fmt"
//	"github.com/ldaidone/gomemo/memo"
//)
//
//func ExampleMemoizeFunc() {
//	m := memo.New()
//
//	slow := func(ctx context.Context, v ...any) (int, error) {
//		return v[0].(int) * v[0].(int), nil
//	}
//
//	memoized := m.MemoizeFunc(func(ctx context.Context, args ...any) (any, error) {
//		return slow(ctx, args)
//	})
//
//	ctx := context.Background()
//	res, _ := memoized(ctx, 3)
//	fmt.Println(res)
//	// Output: 9
//}
