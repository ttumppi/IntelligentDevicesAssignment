package SQLite

import (
	"context"
	"database/sql"
	"goapi/internal/api/repository/DAL"
	"goapi/internal/api/repository/models"
)

type DataRepository struct {
	sqlDB *sql.DB
	createStmt,
	readStmt,
	readManyStmt,
	updateStmt,
	deleteStmt *sql.Stmt
	ctx context.Context
}

func NewDataRepository(sqlDB DAL.SQLDatabase, ctx context.Context) (models.DataRepository, error) {

	repo := &DataRepository{
		sqlDB: sqlDB.Connection(),
		ctx:   ctx,
	}

	// Create the data table if it doesn't exist
	if _, err := repo.sqlDB.Exec(`CREATE TABLE  IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		message VARCHAR(255)
	);`); err != nil {
		repo.sqlDB.Close()
		return nil, err
	}

	// * Create needed Prepared SQL statements, this is more efficient than running each query individually
	createStmt, err := repo.sqlDB.Prepare(`INSERT INTO messages (message) VALUES (?)`)
	if err != nil {
		repo.sqlDB.Close() // Close the database connection if statement preparation fails
		return nil, err
	}
	repo.createStmt = createStmt

	readStmt, err := repo.sqlDB.Prepare("SELECT id, message FROM messages WHERE id = ?")
	if err != nil {
		repo.sqlDB.Close()
		return nil, err
	}
	repo.readStmt = readStmt

	readManyStmt, err := repo.sqlDB.Prepare("SELECT id, message FROM messages LIMIT ? OFFSET ?")
	if err != nil {
		repo.sqlDB.Close()
		return nil, err
	}
	repo.readManyStmt = readManyStmt

	updateStmt, err := repo.sqlDB.Prepare("UPDATE messages SET message = ? WHERE id = ?")
	if err != nil {
		repo.sqlDB.Close()
		return nil, err
	}
	repo.updateStmt = updateStmt

	deleteStmt, err := repo.sqlDB.Prepare("DELETE FROM messages WHERE id = ?")
	if err != nil {
		repo.sqlDB.Close()
		return nil, err
	}
	repo.deleteStmt = deleteStmt

	go Close(ctx, repo)

	return repo, nil
}

func Close(ctx context.Context, r *DataRepository) {

	<-ctx.Done()
	r.createStmt.Close()
	r.readStmt.Close()
	r.updateStmt.Close()
	r.deleteStmt.Close()
	r.readManyStmt.Close()
	r.sqlDB.Close()
}

func (r *DataRepository) Create(data *models.Data, ctx context.Context) error {

	res, err := r.createStmt.ExecContext(ctx, data.DeviceID, data.DeviceName, data.Value, data.Type, data.DateTime, data.Description)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	data.ID = int(id)
	return nil
}

func (r *DataRepository) ReadOne(id int, ctx context.Context) (*models.Data, error) {
	row := r.readStmt.QueryRowContext(ctx, id)
	var data models.Data
	err := row.Scan(&data.ID, &data.DeviceID, &data.DeviceName, &data.Value, &data.Type, &data.DateTime, &data.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &data, nil
}

func (r *DataRepository) ReadMany(page int, rowsPerPage int, ctx context.Context) ([]*models.Data, error) {

	if page < 1 {
		return r.ReadAll()
	}

	offset := rowsPerPage * (page - 1)
	rows, err := r.readManyStmt.QueryContext(ctx, rowsPerPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []*models.Data
	for rows.Next() {
		var d models.Data
		err := rows.Scan(&d.ID, &d.DeviceID, &d.DeviceName, &d.Value, &d.Type, &d.DateTime, &d.Description)
		if err != nil {
			return nil, err
		}
		data = append(data, &d)
	}
	return data, nil
}

func (r *DataRepository) ReadAll() ([]*models.Data, error) {
	rows, err := r.sqlDB.QueryContext(context.Background(), "SELECT id, device_id, device_name, value, data_type, date_time, description FROM data")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []*models.Data
	for rows.Next() {
		var d models.Data
		err := rows.Scan(&d.ID, &d.DeviceID, &d.DeviceName, &d.Value, &d.Type, &d.DateTime, &d.Description)
		if err != nil {
			return nil, err
		}
		data = append(data, &d)
	}
	return data, nil
}

func (r *DataRepository) Update(data *models.Data, ctx context.Context) (int64, error) {
	res, err := r.updateStmt.ExecContext(ctx, data.DeviceID, data.DeviceName, data.Value, data.Type, data.DateTime, data.Description, data.ID)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (r *DataRepository) Delete(data *models.Data, ctx context.Context) (int64, error) {
	res, err := r.deleteStmt.ExecContext(ctx, data.ID)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}
