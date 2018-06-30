package sacloud

import (
	"encoding/json"
	"strings"
)

// MobileGateway モバイルゲートウェイ
type MobileGateway struct {
	*Appliance // アプライアンス共通属性

	Remark   *MobileGatewayRemark   `json:",omitempty"` // リマーク
	Settings *MobileGatewaySettings `json:",omitempty"` // モバイルゲートウェイ設定
}

// MobileGatewayRemark リマーク
type MobileGatewayRemark struct {
	*ApplianceRemarkBase
	// TODO Zone
	//Zone *Resource
}

// MobileGatewaySettings モバイルゲートウェイ設定
type MobileGatewaySettings struct {
	MobileGateway *MobileGatewaySetting `json:",omitempty"` // モバイルゲートウェイ設定リスト
}

// MobileGatewaySetting モバイルゲートウェイ設定
type MobileGatewaySetting struct {
	InternetConnection *MGWInternetConnection `json:",omitempty"` // インターネット接続
	Interfaces         []*MGWInterface        `json:",omitempty"` // インターフェース
	StaticRoutes       []*MGWStaticRoute      `json:",omitempty"` // スタティックルート
}

// MGWInternetConnection インターネット接続
type MGWInternetConnection struct {
	Enabled string `json:",omitempty"`
}

// MGWInterface インターフェース
type MGWInterface struct {
	IPAddress      []string `json:",omitempty"`
	NetworkMaskLen int      `json:",omitempty"`
}

// MGWStaticRoute スタティックルート
type MGWStaticRoute struct {
	Prefix  string `json:",omitempty"`
	NextHop string `json:",omitempty"`
}

// MobileGatewayPlan モバイルゲートウェイプラン
type MobileGatewayPlan int

var (
	// MobileGatewayPlanStandard スタンダードプラン // TODO 正式名称不明なため暫定の名前
	MobileGatewayPlanStandard = MobileGatewayPlan(1)
)

// CreateMobileGatewayValue モバイルゲートウェイ作成用パラメーター
type CreateMobileGatewayValue struct {
	Name        string   // 名称
	Description string   // 説明
	Tags        []string // タグ
	IconID      int64    // アイコン
}

// CreateNewMobileGateway モバイルゲートウェイ作成
func CreateNewMobileGateway(values *CreateMobileGatewayValue, setting *MobileGatewaySetting) (*MobileGateway, error) {

	lb := &MobileGateway{
		Appliance: &Appliance{
			Class:           "mobilegateway",
			propName:        propName{Name: values.Name},
			propDescription: propDescription{Description: values.Description},
			propTags:        propTags{Tags: values.Tags},
			propPlanID:      propPlanID{Plan: &Resource{ID: int64(MobileGatewayPlanStandard)}},
			propIcon: propIcon{
				&Icon{
					Resource: NewResource(values.IconID),
				},
			},
		},
		Remark: &MobileGatewayRemark{
			ApplianceRemarkBase: &ApplianceRemarkBase{
				Switch: &ApplianceRemarkSwitch{
					propScope: propScope{
						Scope: "shared",
					},
				},
				Servers: []interface{}{
					nil,
				},
			},
		},
		Settings: &MobileGatewaySettings{
			MobileGateway: setting,
		},
	}

	return lb, nil
}

// SetPrivateInterface プライベート側NICの接続
func (m *MobileGateway) SetPrivateInterface(ip string, nwMaskLen int) {
	if len(m.Settings.MobileGateway.Interfaces) > 1 {
		m.Settings.MobileGateway.Interfaces[1].IPAddress = []string{ip}
		m.Settings.MobileGateway.Interfaces[1].NetworkMaskLen = nwMaskLen
	} else {
		nic := &MGWInterface{
			IPAddress:      []string{ip},
			NetworkMaskLen: nwMaskLen,
		}
		m.Settings.MobileGateway.Interfaces = append(m.Settings.MobileGateway.Interfaces, nic)
	}
}

// ClearPrivateInterface プライベート側NICの切断
func (m *MobileGateway) ClearPrivateInterface() {
	m.Settings.MobileGateway.Interfaces = []*MGWInterface{nil}
}

// NewMobileGatewayResolver DNS登録用パラメータ作成
func NewMobileGatewayResolver(dns1, dns2 string) *MobileGatewayResolver {
	return &MobileGatewayResolver{
		SimGroup: &MobileGatewaySIMGroup{
			DNS1: dns1,
			DNS2: dns2,
		},
	}
}

// MobileGatewayResolver DNS登録用パラメータ
type MobileGatewayResolver struct {
	SimGroup *MobileGatewaySIMGroup `json:"sim_group,omitempty"`
}

// UnmarshalJSON JSONアンマーシャル(配列、オブジェクトが混在するためここで対応)
func (m *MobileGatewaySIMGroup) UnmarshalJSON(data []byte) error {
	targetData := strings.Replace(strings.Replace(string(data), " ", "", -1), "\n", "", -1)
	if targetData == `[]` {
		return nil
	}

	tmp := &struct {
		DNS1 string `json:"dns_1,omitempty"`
		DNS2 string `json:"dns_2,omitempty"`
	}{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	m.DNS1 = tmp.DNS1
	m.DNS2 = tmp.DNS2
	return nil
}

// MobileGatewaySIMGroup DNS登録用SIMグループ値
type MobileGatewaySIMGroup struct {
	DNS1 string `json:"dns_1,omitempty"`
	DNS2 string `json:"dns_2,omitempty"`
}
