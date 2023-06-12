package iff

import (
	_ "github.com/pangbox/server/common" // for restruct.EnableExprBeta
	"github.com/pangbox/server/pangya"
)

type ItemV11_78 struct {
	/* 0x00 */ Active bool
	/* 0x01 */ _ [3]byte
	/* 0x04 */ ID uint32
	/* 0x08 */ Name string `struct:"[40]byte"`
	/* 0x30 */ Rank byte
	/* 0x31 */ Icon string `struct:"[40]byte"`
	/* 0x59 */ _ [3]byte
	/* 0x5C */ Price uint32
	/* 0x60 */ DiscountPrice uint32
	/* 0x64 */ Condition uint32
	/* 0x68 */ ShopFlag byte
	/* 0x69 */ MoneyFlag byte
	/* 0x6A */ TimeFlag byte
	/* 0x6B */ TimeByte byte
	/* 0x6C */ Quantity uint16
	/* 0x6E */ Unknown2 [4]uint16
	/* 0x76 */ _ [2]byte
	/* 0x78 */
}

type ItemV11_98 struct {
	/* 0x00 */ Active bool
	/* 0x01 */ _ [3]byte
	/* 0x04 */ ID uint32
	/* 0x08 */ Name string `struct:"[40]byte"`
	/* 0x30 */ Rank byte
	/* 0x31 */ Icon string `struct:"[40]byte"`
	/* 0x59 */ _ [3]byte
	/* 0x5C */ Price uint32
	/* 0x60 */ DiscountPrice uint32
	/* 0x64 */ Condition uint32
	/* 0x68 */ ShopFlag byte
	/* 0x69 */ MoneyFlag byte
	/* 0x6A */ TimeFlag byte
	/* 0x6B */ TimeByte byte
	/* 0x6C */ StartTime pangya.SystemTime
	/* 0x7C */ EndTime pangya.SystemTime
	/* 0x6C */ Quantity uint16
	/* 0x6E */ Unknown2 [4]uint16
	/* 0x96 */ _ [2]byte
	/* 0x98 */
}

type ItemV11_B0 struct {
	/* 0x00 */ Active bool
	/* 0x01 */ _ [3]byte
	/* 0x04 */ ID uint32
	/* 0x08 */ Name string `struct:"[64]byte"`
	/* 0x48 */ Rank byte
	/* 0x49 */ Icon string `struct:"[40]byte"`
	/* 0x71 */ _ [3]byte
	/* 0x74 */ Price uint32
	/* 0x78 */ DiscountPrice uint32
	/* 0x7C */ Condition uint32
	/* 0x80 */ ShopFlag byte
	/* 0x81 */ MoneyFlag byte
	/* 0x82 */ TimeFlag byte
	/* 0x83 */ TimeByte byte
	/* 0x84 */ StartTime pangya.SystemTime
	/* 0x94 */ EndTime pangya.SystemTime
	/* 0x6C */ Quantity uint16
	/* 0x6E */ Unknown2 [4]uint16
	/* 0xAE */ _ [2]byte
	/* 0xB0 */
}

type ItemV11_C0 struct {
	/* 0x00 */ Active bool
	/* 0x01 */ _ [3]byte
	/* 0x04 */ ID uint32
	/* 0x08 */ Name string `struct:"[40]byte"`
	/* 0x30 */ Rank byte
	/* 0x31 */ Icon string `struct:"[40]byte"`
	/* 0x59 */ _ [3]byte
	/* 0x5C */ Price uint32
	/* 0x60 */ DiscountPrice uint32
	/* 0x64 */ Condition uint32
	/* 0x68 */ ShopFlag byte
	/* 0x69 */ MoneyFlag byte
	/* 0x6A */ TimeFlag byte
	/* 0x6B */ TimeByte byte
	/* 0x6C */ StartTime pangya.SystemTime
	/* 0x7C */ EndTime pangya.SystemTime
	/* 0x8C */ Model string `struct:"[40]byte"`
	/* 0x6C */ Quantity uint16
	/* 0x6E */ Unknown2 [4]uint16
	/* 0xBE */ _ [2]byte
	/* 0xC0 */
}

type ItemV11_D8_1 struct {
	/* 0x00 */ Active bool
	/* 0x01 */ _ [3]byte
	/* 0x04 */ ID uint32
	/* 0x08 */ Name string `struct:"[64]byte"`
	/* 0x48 */ Rank byte
	/* 0x49 */ Icon string `struct:"[40]byte"`
	/* 0x71 */ _ [3]byte
	/* 0x74 */ Price uint32
	/* 0x78 */ DiscountPrice uint32
	/* 0x7C */ Condition uint32
	/* 0x80 */ ShopFlag byte
	/* 0x81 */ MoneyFlag byte
	/* 0x82 */ TimeFlag byte
	/* 0x83 */ TimeByte byte
	/* 0x84 */ Model string `struct:"[40]byte"`
	/* 0x84 */ StartTime pangya.SystemTime
	/* 0x94 */ EndTime pangya.SystemTime
	/* 0x6C */ Quantity uint16
	/* 0x6E */ Unknown2 [4]uint16
	/* 0xD6 */ _ [2]byte
	/* 0xD8 */
}

type ItemV11_D8_2 struct {
	/* 0x00 */ Active bool
	/* 0x01 */ _ [3]byte
	/* 0x04 */ ID uint32
	/* 0x08 */ Name string `struct:"[64]byte"`
	/* 0x48 */ Rank byte
	/* 0x49 */ Icon string `struct:"[40]byte"`
	/* 0x71 */ _ [3]byte
	/* 0x74 */ Price uint32
	/* 0x78 */ DiscountPrice uint32
	/* 0x7C */ Condition uint32
	/* 0x80 */ ShopFlag byte
	/* 0x81 */ MoneyFlag byte
	/* 0x82 */ TimeFlag byte
	/* 0x83 */ TimeByte byte
	/* 0x** */ StartTime pangya.SystemTime
	/* 0x** */ EndTime pangya.SystemTime
	/* 0x** */ Model string `struct:"[40]byte"`
	/* 0x6C */ Quantity uint16
	/* 0x6E */ Unknown2 [4]uint16
	/* 0xD6 */ _ [2]byte
	/* 0xD8 */
}

type ItemV11_C4 struct {
	/* 0x00 */ Active bool
	/* 0x01 */ _ [3]byte
	/* 0x04 */ ID uint32
	/* 0x08 */ Name string `struct:"[40]byte"`
	/* 0x30 */ Rank byte
	/* 0x31 */ Icon string `struct:"[40]byte"`
	/* 0x59 */ _ [3]byte
	/* 0x5C */ Price uint32
	/* 0x60 */ DiscountPrice uint32
	/* 0x64 */ Condition uint32
	/* 0x68 */ ShopFlag byte
	/* 0x69 */ MoneyFlag byte
	/* 0x6A */ TimeFlag byte
	/* 0x6B */ TimeByte byte
	/* 0x6C */ Unknown uint32
	/* 0x70 */ StartTime pangya.SystemTime
	/* 0x80 */ EndTime pangya.SystemTime
	/* 0x90 */ Model string `struct:"[40]byte"`
	/* 0x6C */ Quantity uint16
	/* 0x6E */ Unknown2 [4]uint16
	/* 0xC2 */ _ [2]byte
	/* 0xC4 */
}

type ItemV13_E0 struct {
	/* 0x00 */ Active bool
	/* 0x01 */ _ [3]byte
	/* 0x04 */ ID uint32
	/* 0x08 */ Name string `struct:"[40]byte"`
	/* 0x30 */ Rank byte
	/* 0x31 */ Icon string `struct:"[40]byte"`
	/* 0x59 */ _ [3]byte
	/* 0x5C */ Price uint32
	/* 0x60 */ DiscountPrice uint32
	/* 0x64 */ Condition uint32
	/* 0x68 */ ShopFlag byte
	/* 0x69 */ MoneyFlag byte
	/* 0x6A */ TimeFlag byte
	/* 0x6B */ TimeByte byte
	/* 0x6C */ Point uint32
	/* 0x70 */ Unknown [0x1C]byte
	/* 0x8C */ StartTime pangya.SystemTime
	/* 0x9C */ EndTime pangya.SystemTime
	/* 0xAC */ Model string `struct:"[40]byte"`
	/* 0x6C */ Quantity uint16
	/* 0x6E */ Unknown2 [4]uint16
	/* 0xDE */ _ [2]byte
	/* 0xE0 */
}

type ItemV13_F8 struct {
	/* 0x00 */ Active bool
	/* 0x01 */ _ [3]byte
	/* 0x04 */ ID uint32
	/* 0x08 */ Name string `struct:"[64]byte"`
	/* 0x48 */ Rank byte
	/* 0x49 */ Icon string `struct:"[40]byte"`
	/* 0x71 */ _ [3]byte
	/* 0x74 */ Price uint32
	/* 0x78 */ DiscountPrice uint32
	/* 0x7C */ Condition uint32
	/* 0x80 */ ShopFlag byte
	/* 0x81 */ MoneyFlag byte
	/* 0x82 */ TimeFlag byte
	/* 0x83 */ TimeByte byte
	/* 0x84 */ Point uint32
	/* 0x88 */ Unknown [0x1C]byte
	/* 0xA4 */ StartTime pangya.SystemTime
	/* 0xB4 */ EndTime pangya.SystemTime
	/* 0xC4 */ Model string `struct:"[40]byte"`
	/* 0x6C */ Quantity uint16
	/* 0x6E */ Unknown2 [4]uint16
	/* 0xF6 */ _ [2]byte
	/* 0xF8 */
}
