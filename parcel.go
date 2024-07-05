package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	/*stmt, err := s.db.Prepare("INSERT INTO parcel (number, client, address, status) VALUES (:number, :client, :address, :status)")
	if err != nil {
		return 0, fmt.Errorf("Не удалось подготовить оператор вставки: %w", err)
	}
	defer stmt.Close()
	// Выполните подготовленный оператор с данными о посылке
	res, err := stmt.Exec(p.Number, p.Client, p.Address, p.Status)
	if err != nil {
		return 0, fmt.Errorf("не удалось выполнить оператор вставки: %w", err)
	}*/
	// Получите идентификатор новой вставленной записи
	res, err := s.db.Exec("INSERT INTO parcel (number, client, address, status, created_at) VALUES (:number, :client, :address, :status, :created_at)",
		sql.Named("number", p.Number),
		sql.Named("client", p.Client),
		sql.Named("address", p.Address),
		sql.Named("status", p.Status),
		sql.Named("created_at", p.CreatedAt))
	if err != nil {
		return 0, err
	}
	// Проверка на дубликат
	var exist int
	err = s.db.QueryRow("SELECT 1 FROM parcel WHERE number = ?", p.Number).Scan(&exist)
	if err == nil {
		return 0, fmt.Errorf("посылка с таким номером уже существует:%w", err)
	}
	lastInsertID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("не удалось получить идентификатор последней вставки: %w", err)
	}
	// верните идентификатор последней добавленной записи
	return int(lastInsertID), nil
	// return 0, nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	stmt, err := s.db.Prepare("SELECT number, client, address, status, created_at FROM parcel WHERE number = ?")
	if err != nil {
		return Parcel{}, fmt.Errorf("не удалось подготовить оператор select: %w", err)
	}
	defer stmt.Close()
	// Выполните подготовленный оператор с номером участка
	row := stmt.QueryRow(number)
	// Создайте объект Parcel для хранения полученных данных
	p := Parcel{}
	// var p Parcel
	// Сканируйте значения из строки в поля объекта Parcel
	err = row.Scan(&p.Number, &p.Client, &p.Address, &p.Status, &p.CreatedAt) //&p.CreatedAt)
	if err != nil {
		return Parcel{}, fmt.Errorf("не удалось отсканировать строку: %w", err)
	}
	// заполните объект Parcel данными из таблицы
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	stmt, err := s.db.Prepare("SELECT number, client, address, status, created_at FROM parcel WHERE client = ?")
	if err != nil {
		return nil, fmt.Errorf("не удалось подготовить оператор select: %w", err)
	}
	defer stmt.Close()
	// Выполните подготовленный оператор с идентификатором клиента
	rows, err := stmt.Query(client)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить оператор select: %w", err)
	}
	defer rows.Close()
	// Создайте фрагмент для хранения полученных посылок
	var parcell []Parcel
	// Итерация по строкам, возвращенным запросом
	for rows.Next() {
		// Создаем объект Parcel для хранения данных из каждого ряда
		var p Parcel
		// Сканируем значения из строки в поля объекта Parcel
		err = rows.Scan(&p.Number, &p.Client, &p.Address, &p.Status, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("не удалось отсканировать строку: %w", err)
		}
		// Добавить объект Parcel к фрагменту parcels
		parcell = append(parcell, p)
	}
	// Проверка на наличие ошибок во время итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации строк: %w", err)
	}
	return parcell, nil
	// заполните срез Parcel данными из таблицы
	// var res []Parcel
	// return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	stmt, err := s.db.Prepare("UPDATE parcel SET status = ? WHERE number = ?")
	if err != nil {
		return fmt.Errorf("не удалось подготовить оператор обновления: %w", err)
	}
	defer stmt.Close()
	// Выполните подготовленный оператор с новым статусом и номером участка
	_, err = stmt.Exec(status, number)
	if err != nil {
		return fmt.Errorf("не удалось выполнить оператор обновления: %w", err)
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	stmt, err := s.db.Prepare("UPDATE parcel SET address = ? WHERE number = ? AND status = ?")
	if err != nil {
		return fmt.Errorf("не удалось подготовить оператор обновления: %w", err)
	}
	defer stmt.Close()
	// Выполните подготовленный оператор с новым адресом, номером участка и статусом
	_, err = stmt.Exec(address, number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("не удалось выполнить оператор обновления: %w", err)
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	stmt, err := s.db.Prepare("DELETE FROM parcel WHERE number = ? AND status = ?")
	if err != nil {
		return fmt.Errorf("не удалось подготовить оператор удаления: %w", err)
	}
	defer stmt.Close()
	// Выполните подготовленный оператор с номером и статусом участка
	_, err = stmt.Exec(number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("не удалось выполнить оператор delete: %w", err)
	}
	return nil
}
