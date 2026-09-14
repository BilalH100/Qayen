package utils

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

func Float8ToString(f pgtype.Float8) string {
	if !f.Valid {
		return "NULL"
	}
	return fmt.Sprintf("%f", f.Float64)
}

func StringToFloat8(s string) (pgtype.Float8, error) {
	var f pgtype.Float8
	if s == "" {
		f.Valid = false
		return f, nil
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return f, err
	}
	f.Float64 = val
	f.Valid = true
	return f, nil
}
func Int32ToPgInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: i,
		Valid: true,
	}
}
