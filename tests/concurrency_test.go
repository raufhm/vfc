package tests

import (
	"fmt"
	"sync"
	"testing"

	"github.com/raufhm/vfc/internal/domain"
	"github.com/raufhm/vfc/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentWrites(t *testing.T) {
	repo := repository.NewInMemoryRepository()

	var wg sync.WaitGroup
	numWorkers := 50

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				productID := fmt.Sprintf("product-%d", j)
				product := domain.NewProduct(productID, float64(id*10+j), id*10+j)
				err := repo.Save(product)
				require.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		productID := fmt.Sprintf("product-%d", i)
		product, err := repo.Get(productID)
		require.NoError(t, err)
		assert.Equal(t, productID, product.ProductID)
	}
}

func TestConcurrentReads(t *testing.T) {
	repo := repository.NewInMemoryRepository()

	for i := 0; i < 10; i++ {
		productID := fmt.Sprintf("product-%d", i)
		product := domain.NewProduct(productID, float64(i)*10, i*100)
		err := repo.Save(product)
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	numReaders := 50

	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				productID := fmt.Sprintf("product-%d", j)
				product, err := repo.Get(productID)
				require.NoError(t, err)
				assert.Equal(t, productID, product.ProductID)
			}
		}()
	}

	wg.Wait()
}

func TestMixedReadWrite(t *testing.T) {
	repo := repository.NewInMemoryRepository()

	for i := 0; i < 10; i++ {
		productID := fmt.Sprintf("product-%d", i)
		product := domain.NewProduct(productID, float64(i)*10, i*100)
		err := repo.Save(product)
		require.NoError(t, err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				productID := fmt.Sprintf("product-%d", j%10)
				product := domain.NewProduct(productID, float64(id*20+j), id*20+j)
				err := repo.Save(product)
				require.NoError(t, err)
			}
		}(i)
	}

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				productID := fmt.Sprintf("product-%d", j%10)
				product, err := repo.Get(productID)
				require.NoError(t, err)
				assert.NotNil(t, product)
			}
		}()
	}

	wg.Wait()
}
