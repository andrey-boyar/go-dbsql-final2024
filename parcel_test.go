package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Number:    randRange.Intn(1000), // Генерируем случайный номер
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close() // настройте подключение к БД

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки")
	assert.True(t, id > 0, "Идентификатор должен быть больше нуля")
	// assert.Equal(t, parcel, id, "Ошибка")

	// get 	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	retrievedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка при получении посылки")
	assert.Equal(t, parcel, retrievedParcel, "Ошибка")

	// delete	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err, "Ошибка при удалении посылки")
	// check
	_, err = store.Get(id)
	assert.Error(t, err, "Ошибка")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close() // настройте подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add 	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	insertID, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки")
	assert.True(t, insertID > 0, "Идентификатор должен быть больше нуля")

	// set address	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(insertID, newAddress)
	require.NoError(t, err, "Ошибка установки адреса")

	// check	// получите добавленную посылку и убедитесь, что адрес обновился
	retrievedParcel, err := store.Get(insertID)
	require.NoError(t, err, "Ошибка при получении посылки")
	assert.Equal(t, newAddress, retrievedParcel.Address, "Адрес неправильный")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close() // настройте подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	insertID, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки")
	assert.True(t, insertID > 0, "Идентификатор должен быть больше нуля")

	// set status	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(insertID, ParcelStatusSent)
	require.NoError(t, err, "Ошибка установки статуса")

	// check	// получите добавленную посылку и убедитесь, что статус обновился
	retrievedParcel, err := store.Get(insertID)
	require.NoError(t, err, "Ошибка при получении посылки")
	assert.Equal(t, ParcelStatusSent, retrievedParcel.Status, "Статус не корректный")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close() // настройте подключение к БД

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := make(map[int]Parcel) //{}
	store := ParcelStore{db: db}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "Ошибка при добавлении посылки")
		assert.True(t, id > 0, "Идентификатор должен быть больше нуля")
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id
		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "Ошибка при получении посылок клиентом")
	assert.Equal(t, len(parcels), len(storedParcels), "Количество полученных посылок должно совпадать с количеством добавленных посылок")
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		expectedParcel, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		assert.Equal(t, expectedParcel, parcel, "Полученная посылка не соответствует ожидаемым значениям")
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
