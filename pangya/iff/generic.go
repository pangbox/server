package iff

import "github.com/pangbox/server/pangya"

type Common struct {
}

type Item struct {
	Active        bool
	ID            uint32
	Name          string
	MaxRank       byte
	Icon          string
	Price         uint32
	DiscountPrice uint32
	Condition     uint32
	ShopFlag      byte
	MoneyFlag     byte
	TimeFlag      byte
	TimeByte      byte
	Point         uint32
	StartTime     pangya.SystemTime
	EndTime       pangya.SystemTime
	Common        Common
	Model         string
	Quantity      uint16
	Unknown2      [4]uint16
}

type itemGeneric interface {
	Generic() Item
}

func (i ItemV11_78) Generic() Item {
	return Item{
		Active:        i.Active,
		ID:            i.ID,
		Name:          i.Name,
		MaxRank:       i.MaxRank,
		Icon:          i.Icon,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Condition:     i.Condition,
		ShopFlag:      i.ShopFlag,
		MoneyFlag:     i.MoneyFlag,
		TimeFlag:      i.TimeFlag,
		TimeByte:      i.TimeByte,
		Model:         i.Icon,
		Quantity:      i.Quantity,
		Unknown2:      i.Unknown2,
	}
}

func (i ItemV11_98) Generic() Item {
	return Item{
		Active:        i.Active,
		ID:            i.ID,
		Name:          i.Name,
		MaxRank:       i.MaxRank,
		Icon:          i.Icon,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Condition:     i.Condition,
		ShopFlag:      i.ShopFlag,
		MoneyFlag:     i.MoneyFlag,
		TimeFlag:      i.TimeFlag,
		TimeByte:      i.TimeByte,
		Model:         i.Icon,
		Quantity:      i.Quantity,
		Unknown2:      i.Unknown2,
	}
}

func (i ItemV11_B0) Generic() Item {
	return Item{
		Active:        i.Active,
		ID:            i.ID,
		Name:          i.Name,
		MaxRank:       i.MaxRank,
		Icon:          i.Icon,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Condition:     i.Condition,
		ShopFlag:      i.ShopFlag,
		MoneyFlag:     i.MoneyFlag,
		TimeFlag:      i.TimeFlag,
		TimeByte:      i.TimeByte,
		StartTime:     i.StartTime,
		EndTime:       i.EndTime,
		Model:         i.Icon,
		Quantity:      i.Quantity,
		Unknown2:      i.Unknown2,
	}
}

func (i ItemV11_C0) Generic() Item {
	return Item{
		Active:        i.Active,
		ID:            i.ID,
		Name:          i.Name,
		MaxRank:       i.MaxRank,
		Icon:          i.Icon,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Condition:     i.Condition,
		ShopFlag:      i.ShopFlag,
		MoneyFlag:     i.MoneyFlag,
		TimeFlag:      i.TimeFlag,
		TimeByte:      i.TimeByte,
		StartTime:     i.StartTime,
		EndTime:       i.EndTime,
		Model:         i.Model,
		Quantity:      i.Quantity,
		Unknown2:      i.Unknown2,
	}
}

func (i ItemV11_D8_1) Generic() Item {
	return Item{
		Active:        i.Active,
		ID:            i.ID,
		Name:          i.Name,
		MaxRank:       i.MaxRank,
		Icon:          i.Icon,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Condition:     i.Condition,
		ShopFlag:      i.ShopFlag,
		MoneyFlag:     i.MoneyFlag,
		TimeFlag:      i.TimeFlag,
		TimeByte:      i.TimeByte,
		StartTime:     i.StartTime,
		EndTime:       i.EndTime,
		Model:         i.Model,
		Quantity:      i.Quantity,
		Unknown2:      i.Unknown2,
	}
}

func (i ItemV11_D8_2) Generic() Item {
	return Item{
		Active:        i.Active,
		ID:            i.ID,
		Name:          i.Name,
		MaxRank:       i.MaxRank,
		Icon:          i.Icon,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Condition:     i.Condition,
		ShopFlag:      i.ShopFlag,
		MoneyFlag:     i.MoneyFlag,
		TimeFlag:      i.TimeFlag,
		TimeByte:      i.TimeByte,
		StartTime:     i.StartTime,
		EndTime:       i.EndTime,
		Model:         i.Model,
		Quantity:      i.Quantity,
		Unknown2:      i.Unknown2,
	}
}

func (i ItemV13_E0) Generic() Item {
	return Item{
		Active:        i.Active,
		ID:            i.ID,
		Name:          i.Name,
		MaxRank:       i.MaxRank,
		Icon:          i.Icon,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Condition:     i.Condition,
		ShopFlag:      i.ShopFlag,
		MoneyFlag:     i.MoneyFlag,
		TimeFlag:      i.TimeFlag,
		TimeByte:      i.TimeByte,
		Point:         i.Point,
		StartTime:     i.StartTime,
		EndTime:       i.EndTime,
		Model:         i.Model,
		Quantity:      i.Quantity,
		Unknown2:      i.Unknown2,
	}
}

func (i ItemV11_C4) Generic() Item {
	return Item{
		Active:        i.Active,
		ID:            i.ID,
		Name:          i.Name,
		MaxRank:       i.MaxRank,
		Icon:          i.Icon,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Condition:     i.Condition,
		ShopFlag:      i.ShopFlag,
		MoneyFlag:     i.MoneyFlag,
		TimeFlag:      i.TimeFlag,
		TimeByte:      i.TimeByte,
		StartTime:     i.StartTime,
		EndTime:       i.EndTime,
		Model:         i.Model,
		Quantity:      i.Quantity,
		Unknown2:      i.Unknown2,
	}
}

func (i ItemV13_F8) Generic() Item {
	return Item{
		Active:        i.Active,
		ID:            i.ID,
		Name:          i.Name,
		MaxRank:       i.MaxRank,
		Icon:          i.Icon,
		Price:         i.Price,
		DiscountPrice: i.DiscountPrice,
		Condition:     i.Condition,
		ShopFlag:      i.ShopFlag,
		MoneyFlag:     i.MoneyFlag,
		TimeFlag:      i.TimeFlag,
		TimeByte:      i.TimeByte,
		Point:         i.Point,
		StartTime:     i.StartTime,
		EndTime:       i.EndTime,
		Model:         i.Model,
		Quantity:      i.Quantity,
		Unknown2:      i.Unknown2,
	}
}
