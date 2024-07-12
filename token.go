// copied from https://gist.github.com/rishabhpoddar/8c26ed237add1a5b86481e72032abf8d
package shield

import (
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
)

var jwksCache *sessmodels.GetJWKSResult = nil
var mutex sync.RWMutex
var JWKCacheMaxAgeInMs int64 = 60000
var coreUrl = "https://try.supertokens.com/.well-known/jwks.json"

func getJWKSFromCacheIfPresent() *sessmodels.GetJWKSResult {
	mutex.RLock()
	defer mutex.RUnlock()
	if jwksCache != nil {
		// This means that we have valid JWKs for the given core path
		// We check if we need to refresh before returning
		currentTime := time.Now().UnixNano() / int64(time.Millisecond)

		// This means that the value in cache is not expired, in this case we return the cached value
		//
		// Note that this also means that the SDK will not try to query any other Core (if there are multiple)
		// if it has a valid cache entry from one of the core URLs. It will only attempt to fetch
		// from the cores again after the entry in the cache is expired
		if (currentTime - jwksCache.LastFetched) < JWKCacheMaxAgeInMs {
			return jwksCache
		}
	}

	return nil
}

func GetJWKS() (*keyfunc.JWKS, error) {
	resultFromCache := getJWKSFromCacheIfPresent()

	if resultFromCache != nil {
		return resultFromCache.JWKS, nil
	}

	mutex.Lock()
	defer mutex.Unlock()
	// RefreshUnknownKID - Fetch JWKS again if the kid in the header of the JWT does not match any in
	// the keyfunc library's cache
	jwks, err := keyfunc.Get(coreUrl, keyfunc.Options{
		RefreshUnknownKID: true,
	})
	if err != nil {
		return nil, err
	}

	jwksResult := sessmodels.GetJWKSResult{
		JWKS:        jwks,
		Error:       nil,
		LastFetched: time.Now().UnixNano() / int64(time.Millisecond),
	}

	// Dont add to cache if there is an error to keep the logic of checking cache simple
	//
	// This also has the added benefit where if initially the request failed because the core
	// was down and then it comes back up, the next time it will try to request that core again
	// after the cache has expired
	jwksCache = &jwksResult

	return jwksResult.JWKS, nil
}
