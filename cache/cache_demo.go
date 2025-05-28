package cache

import (
	"log"
	"sync"
)

// 外部数据结构
type ClothesData struct {
	ID         int64
	ActiveList map[int32]int64        // 激活衣柜列表
	MiscData   map[string]interface{} // 数据杂项
	mutex      sync.RWMutex           // 独立锁
}

// 假设你有一个自定义的 handler 需要落库
type PlayerClothesFlushHandler struct{}

func isClothesData(v interface{}) (ClothesData, bool) {
	switch v := v.(type) {
	case ClothesData:
		return v, true
	case *ClothesData:
		return *v, true
	default:
		return ClothesData{}, false
	}
}

// 批量落地sql语句
func BatchUpdatePlayersClothes(updates []ClothesData) error {
	// if len(updates) == 0 {
	// 	return nil
	// }

	// tx := query.Q.Begin()
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		tx.Rollback()
	// 	}
	// }()

	// var sqlBuilder strings.Builder
	// var params []interface{}
	// ids := make([]int64, 0, len(updates))

	// // Update car_id and car_list
	// sqlBuilder.WriteString("UPDATE tb_clothes SET active_list = CASE player_id ")
	// for _, update := range updates {
	// 	sqlBuilder.WriteString("WHEN ? THEN ? ")
	// 	params = append(params, update.ID, gamepack.PackJsonAny(update.ActiveList))
	// 	ids = append(ids, update.ID)
	// }

	// sqlBuilder.WriteString("END, misc_info = CASE player_id ")
	// for _, update := range updates {
	// 	sqlBuilder.WriteString("WHEN ? THEN ? ")
	// 	params = append(params, update.ID, gamepack.PackJsonAny(update.MiscData))
	// }

	// sqlBuilder.WriteString("END WHERE player_id IN (?" + strings.Repeat(",?", len(ids)-1) + ")")
	// for _, id := range ids {
	// 	params = append(params, id)
	// }

	// fmt.Println(".................", sqlBuilder.String(), params)

	// result := tx.DB().Exec(sqlBuilder.String(), params...)
	// if err := result.Error; err != nil {
	// 	tx.Rollback()
	// 	return fmt.Errorf("failed to execute batch update: %w", err)
	// }

	// if rowsAffected := result.RowsAffected; int(rowsAffected) != len(ids) {
	// 	log.Printf("expected to update %d rows, but only %d were updated", len(ids), rowsAffected)
	// }

	// if err := tx.Commit(); err != nil {
	// 	tx.Rollback()
	// 	return fmt.Errorf("failed to commit transaction: %w", err)
	// }

	// return nil
	return nil
}

func (h *PlayerClothesFlushHandler) Flush(data []interface{}) error {
	// 处理 Flush 操作
	dataLen := len(data)
	if dataLen == 0 {
		return nil
	}
	dataList := make([]ClothesData, 0, dataLen) // 创建一个空的 []Player 切片

	for _, v := range data {
		if data, ok := isClothesData(v); ok {
			dataList = append(dataList, data)
		} else {
			log.Println("Failed to update PlayerDayData data:", data)
		}
	}
	sqlErr := BatchUpdatePlayersClothes(dataList)
	if sqlErr != nil {
		log.Println("Failed to assert value to Player:", sqlErr)
	}
	return nil
}

// 调整为列表记录
type ListHandler struct{}

func (h *ListHandler) Flush(data []interface{}) error {
	// 这里实现你的刷新逻辑，比如将数据写入数据库、发送到远程服务等
	return nil
}
