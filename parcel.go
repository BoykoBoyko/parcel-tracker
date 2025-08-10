package main

import (
	"database/sql"
)

// Хранилище посылок
type ParcelStore struct {
	db *sql.DB
}

// Конструктор + автоинициализация схемы БД (создаём таблицу, если её нет)
func NewParcelStore(db *sql.DB) ParcelStore {
	s := ParcelStore{db: db}
	if err := s.initSchema(); err != nil {
		// в простом учебном проекте — паникуем, чтобы сразу увидеть проблему
		panic(err)
	}
	return s
}

// Создание таблицы, если её ещё нет
func (s ParcelStore) initSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS parcel (
			number     INTEGER PRIMARY KEY AUTOINCREMENT,
			client     INTEGER NOT NULL,
			status     TEXT    NOT NULL,
			address    TEXT    NOT NULL,
			created_at TEXT    NOT NULL
		);
	`)
	return err
}

// Добавить посылку и вернуть её номер (AUTOINCREMENT id).
func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(
		`INSERT INTO parcel (client, status, address, created_at)
		 VALUES (:client, :status, :address, :created_at)`,
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt),
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Получить посылку по номеру
func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	row := s.db.QueryRow(
		`SELECT number, client, status, address, created_at
		   FROM parcel
		  WHERE number = :number`,
		sql.Named("number", number),
	)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return p, err
	}
	return p, nil
}

// Получить все посылки клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at
		   FROM parcel
		  WHERE client = :client`,
		sql.Named("client", client),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Parcel
	for rows.Next() {
		p := Parcel{}
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// Обновить статус
func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec(
		`UPDATE parcel SET status = :status WHERE number = :number`,
		sql.Named("status", status),
		sql.Named("number", number),
	)
	return err
}

// Обновить адрес (разрешено только в статусе "registered")
func (s ParcelStore) SetAddress(number int, address string) error {
	_, err := s.db.Exec(
		`UPDATE parcel
		    SET address = :address
		  WHERE number = :number AND status = :status`,
		sql.Named("address", address),
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered),
	)
	return err
}

// Удалить посылку (разрешено только в статусе "registered")
func (s ParcelStore) Delete(number int) error {
	_, err := s.db.Exec(
		`DELETE FROM parcel
		  WHERE number = :number AND status = :status`,
		sql.Named("number", number),
		sql.Named("status", ParcelStatusRegistered),
	)
	return err
}
