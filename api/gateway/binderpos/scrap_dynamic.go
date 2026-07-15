package binderpos

import (
	"context"

	"mtg-price-checker-sg/gateway"
)

func (i impl) scrapDynamic(ctx context.Context, scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	releaseDynamicProxy, err := gateway.AcquireDynamicProxySlot(ctx)
	if err != nil {
		return nil, err
	}
	defer releaseDynamicProxy()

	return i.scrapWithCollectorFactory(
		ctx,
		scrapVariant,
		storeName,
		baseUrl,
		searchUrl,
		searchStr,
		newDynamicNoRetryCollector,
	)
}
