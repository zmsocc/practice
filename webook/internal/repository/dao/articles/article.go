package articles

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type articleDao struct {
	db *gorm.DB
}

func NewArticleDao(db *gorm.DB) ArticleDAO {
	return &articleDao{
		db: db,
	}
}

func (d *articleDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := d.db.WithContext(ctx).Create(&art).Error
	// 返回自增主键
	return art.Id, err
}

func (d *articleDao) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := d.db.Model(&Article{}).WithContext(ctx).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		})
	err := res.Error
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (d *articleDao) Sync(ctx context.Context, art Article) (int64, error) {
	tx := d.db.WithContext(ctx).Begin()
	now := time.Now().UnixMilli()
	defer func() {
		if tx.Error != nil {
			tx.Rollback()
		}
	}()
	txDAO := NewArticleDao(tx)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = txDAO.Insert(ctx, art)
	} else {
		err = txDAO.UpdateById(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	publishArt := PublishedArticle(art)
	publishArt.Ctime = now
	publishArt.Utime = now
	err = tx.Clauses(clause.OnConflict{
		// ID 冲突的时候。实际上，在 MYSQL 里面写不写都可以
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		}),
	}).Create(&publishArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, tx.Commit().Error
}

func (d *articleDao) SyncStatus(ctx context.Context, id, author int64, status uint8) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", id, author).
			Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		res = tx.Model(&PublishedArticle{}).
			Where("id = ? AND author_id = ?", id, author).
			Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		return nil
	})
}

func (d *articleDao) FindByAuthor(ctx context.Context, uid int64, offset, limit int) ([]Article, error) {
	var arts []Article
	err := d.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", uid).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (d *articleDao) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	return art, err
}

func (d *articleDao) GetPubById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	return art, err
}
