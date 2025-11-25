package inmemory_cache

import (
	"fmt"
	"hash/fnv"
	"log"
	"parser/internal/domain/models"
	"time"
)

// конструктор для создания кэша с указаным количеством шардов и интервалом очистки кэша
func NewInmemoryShardedCache(numShards int, TTL time.Duration) *InmemoryShardedCache {
	// инициализируем базовую структуру кэша

	cache := &InmemoryShardedCache{
		shards:    make([]*Shard, numShards), // инициализируем слайс указателй на шарды
		numShards: numShards,                 // указывавем количество шардов
		stopChan:  make(chan bool),           // инициализируем канал для остановки
	}

	// для каждого шарда инициализируем внутреннюю мапу
	for i := 0; i < numShards; i++ {
		cache.shards[i] = &Shard{
			Items: map[string]CashItem{},
		}
	}

	// асинхронно запускаем метод очистки кэша через поределённый интервал времени
	go cache.cleanUp(TTL)

	return cache
}

// метод получения занчения из кэша по заданному ключу (это хэшированный запрос поиска)
// чтобы реализовать этот метод - нужна функция, которая будет находить нужный шард по заданному ключу (внутри будет хэш-функция)
// результатом будет значение в CashItem и флаг

func (c *InmemoryShardedCache) GetItem(key string) ([]models.SearchResult, bool) {
	// получаем необходимый шард
	shard := c.GetShard(key)
	now := time.Now()
	// лочимся на чтение, так как читаем из мапы
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	val, ok := shard.Items[key] // проверяем наличие значения по ключу в мапе шарда
	if !ok {
		fmt.Printf("No data in cache with key:%s\n", key)
		return nil, false
	}
	// проверяем, не истёк ли TTL у значения
	if now.After(val.expTime) {
		fmt.Printf("Data in cache with key:%s are not valid\n", key)
		return nil, false
	}
	return val.value, true
}

// метод, чтобы находить нужный шард по заданному ключу
func (c *InmemoryShardedCache) GetShard(key string) *Shard {
	// создаём экземпляр хэша
	hashf := fnv.New32a()
	//записываем в хэш наш ключ в виде байтового среза
	_, err := hashf.Write([]byte(key))
	if err != nil {
		log.Println(err.Error())
	}
	//вычисляем индекс нужного нам шарда по ключу
	// если мы хэш по ключу % количество шардов = индекс шарда в диапазоне от 0 до shardNum-1
	// для каждого ключа будет вой шард, там мы будем рапределять данные по шардам
	shardIndex := int(hashf.Sum32()) % c.numShards

	return c.shards[shardIndex]
}

// метод, чтобы записать значение в кэш с заданным TTL
func (c *InmemoryShardedCache) AddItemWithTTL(key string, value []models.SearchResult, ttl time.Duration) {
	// получаем необходимый шард
	shard := c.GetShard(key)
	now := time.Now()

	// берём лок на запись, так как обращаемся к мапе
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.Items[key] = CashItem{
		value:   value,
		expTime: now.Add(ttl), // время жизни для нового занчения - высчитывавем: время на момоент вызова функции + ttl
	}
}
