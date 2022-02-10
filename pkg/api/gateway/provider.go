package gateway

import (
	"encoding/json"
	"github.com/ddx2x/oilmont/pkg/common"
	"github.com/ddx2x/oilmont/pkg/resource/system"
	ucfg "github.com/ddx2x/oilmont/pkg/resource/userconfig"
)

func (gw *Gateway) allowedProviders(tenant string, cfg *ucfg.Config) error {
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]*ucfg.Provider)
	}
	for _, f := range []func(tenant string, cfg *ucfg.Config) error{
		gw.provider, gw.region, gw.az, gw.cluster, gw.namespace} {
		if err := f(tenant, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (gw *Gateway) getProvider(providerName string) (*system.Provider, error) {
	provider := &system.Provider{}
	err := gw.stage.GetByFilter(
		common.DefaultDatabase, common.PROVIDER, provider,
		map[string]interface{}{
			"spec.local_name": providerName,
		},
		true,
	)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func (gw *Gateway) provider(tenant string, cfg *ucfg.Config) (err error) {
	switch {
	case cfg.IsAdmin():
		var providers []system.Provider
		err = gw.stage.ListToObject(common.DefaultDatabase, common.PROVIDER, nil, &providers, true)
		if err != nil {
			return
		}
		for _, provider := range providers {
			cfg.Providers[provider.GetName()] =
				&ucfg.Provider{
					Name:       provider.GetName(),
					ThirdParty: provider.Spec.ThirdParty,
					LocalName:  provider.Spec.LocalName,
				}
		}
		return
	default:
		getTenant, err := gw.getTenant(tenant)
		if err != nil {
			return err
		}
		for pName, _ := range getTenant.Spec.Allowed {
			provider, err := gw.getProvider(pName)
			if err != nil {
				continue
			}
			cfg.Providers[pName] =
				&ucfg.Provider{
					Name:       provider.GetName(),
					ThirdParty: provider.Spec.ThirdParty,
					LocalName:  provider.Spec.LocalName,
				}
		}
	}

	return
}

func (gw *Gateway) getRegion(regionLocalName, providerName string) (*system.Region, error) {
	region := &system.Region{}
	filter := map[string]interface{}{
		"spec.local_name":    regionLocalName,
		"metadata.namespace": providerName,
	}
	err := gw.stage.GetByFilter(common.DefaultDatabase, common.REGION, region, filter, true)
	if err != nil {
		return nil, err
	}
	return region, nil
}

func (gw *Gateway) region(tenant string, cfg *ucfg.Config) (err error) {
	switch {
	case cfg.IsAdmin():
		for provider, _ := range cfg.Providers {
			filter := map[string]interface{}{"metadata.namespace": provider}
			var regions []system.Region
			err = gw.stage.ListToObject(common.DefaultDatabase, common.REGION, filter, &regions, true)
			if err != nil {
				return
			}
			providerStruct := cfg.Providers[provider]
			if providerStruct.Regions == nil {
				providerStruct.Regions = make([]*ucfg.Region, 0)
			}

			for _, region := range regions {
				providerStruct.Regions = append(
					providerStruct.Regions,
					&ucfg.Region{
						Name:      region.Spec.ID,
						LocalName: region.Spec.LocalName,
					},
				)
			}
		}
	default:
		getTenant, err := gw.getTenant(tenant)
		if err != nil {
			return err
		}
		for pName, region := range getTenant.Spec.Allowed {
			regionMap, ok := region.(map[string]interface{})
			if !ok {
				continue
			}

			cfgProvider := cfg.Providers[pName]
			if cfgProvider.Regions == nil {
				cfgProvider.Regions = make([]*ucfg.Region, 0)
			}

			for rLocalName, _ := range regionMap {
				region, err := gw.getRegion(rLocalName, cfgProvider.Name)
				if err != nil {
					continue
				}
				cfgProvider.Regions = append(
					cfgProvider.Regions,
					&ucfg.Region{
						Name:      region.Spec.ID,
						LocalName: region.Spec.LocalName,
					},
				)
			}
		}
	}

	return
}

func (gw *Gateway) getAZ(pName, rName, aLocalName string) (*system.AvailableZone, error) {
	az := &system.AvailableZone{}
	filter := map[string]interface{}{
		"spec.local_name":    aLocalName,
		"metadata.namespace": pName,
		"spec.region":        rName,
	}
	err := gw.stage.GetByFilter(common.DefaultDatabase, common.AVAILABLEZONE, az, filter, true)
	if err != nil {
		return nil, err
	}
	return az, nil
}

func unmarshalList(tag interface{}, data interface{}) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, tag)
}

func (gw *Gateway) az(tenant string, cfg *ucfg.Config) (err error) {
	switch {
	case cfg.IsAdmin():
		for provider, _ := range cfg.Providers {
			cfgProvider := cfg.Providers[provider]
			for _, region := range cfgProvider.Regions {
				filter := map[string]interface{}{
					"spec.region":        region.Name,
					"metadata.namespace": provider,
				}
				var azs []system.AvailableZone
				err = gw.stage.ListToObject(common.DefaultDatabase, common.AVAILABLEZONE, filter, &azs, true)
				if err != nil {
					return
				}
				for _, az := range azs {
					region.Azs = append(region.Azs,
						&ucfg.Az{
							Name:      az.Spec.ID,
							LocalName: az.Spec.LocalName,
						},
					)
				}
			}
		}
		return
	default:
		getTenant, err := gw.getTenant(tenant)
		if err != nil {
			return err
		}
		for pName, region := range getTenant.Spec.Allowed {
			allowedRegionMap, ok := region.(map[string]interface{})
			if !ok {
				continue
			}
			cfgProvider := cfg.Providers[pName]
			for allowedRegionName, allowedAzsInterface := range allowedRegionMap {
				cfgRegion := cfgProvider.GetRegion(allowedRegionName)
				if cfgRegion == nil {
					continue
				}
				var allowedAZs []string
				if err := unmarshalList(&allowedAZs, allowedAzsInterface); err != nil {
					continue
				}
				for _, allowedAZ := range allowedAZs {
					az, err := gw.getAZ(cfgProvider.Name, cfgRegion.Name, allowedAZ)
					if err != nil {
						continue
					}
					cfgRegion.Azs = append(cfgRegion.Azs,
						&ucfg.Az{
							Name:      az.Spec.ID,
							LocalName: az.Spec.LocalName,
						},
					)
				}
			}
		}
	}

	return
}

func (gw *Gateway) cluster(tenant string, cfg *ucfg.Config) (err error) {
	// TODO: 需要写入管理员授权可访问的集群（AZ->Cluster)
	return
}

func (gw *Gateway) namespace(tenant string, cfg *ucfg.Config) (err error) {
	// TODO: 用户自主开通区域的资源池(Namespace)
	return
}
