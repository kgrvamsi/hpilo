package redfishapi

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	ver "github.com/hashicorp/go-version"
)

// Declaring the Constant Values
const (
	StatusUnauthorized        = "Unauthorized"
	StatusInternalServerError = "Server Error"
	StatusBadRequest          = "Bad Request"
)

//StartServerDell ...
// ResetType@Redfish.AllowableValues
// 0	"On"
// 1	"ForceOff"
// 2	"GracefulRestart"
// 3	"GracefulShutdown"
// 4	"PushPowerButton"
// 5	"Nmi"
// target: "/redfish/v1/Systems/System.Embedded.1/Actions/ComputerSystem.Reset"
// works: R730xd,R740xd
func (c *IloClient) StartServerDell() (string, error) {
	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1/Actions/ComputerSystem.Reset"

	var jsonStr = []byte(`{"ResetType": "On"}`)
	_, _, err := queryData(c, "POST", url, jsonStr)
	if err != nil {
		return "", err
	}

	return "Server Started", nil
}

//StopServerDell ... Will Request to stop the server
// works: R730xd,R740xd
func (c *IloClient) StopServerDell() (string, error) {
	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1/Actions/ComputerSystem.Reset"

	var jsonStr = []byte(`{"ResetType": "ForceOff"}`)
	_, _, err := queryData(c, "POST", url, jsonStr)
	if err != nil {
		return "", err
	}

	return "Server Stopped", nil
}

//GracefulRestartDell ... Will Reset Idrac and will take some time to come up
func (c *IloClient) GracefulRestartDell() (string, error) {
	url := c.Hostname + "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Manager.Reset"

	var jsonStr = []byte(`{"ResetType": "GracefulRestart"}`)
	_, _, err := queryData(c, "POST", url, jsonStr)
	if err != nil {
		return "", err
	}

	return "Idrac Reset", nil

}

//GetServerPowerStateDell ... Will fetch the current state of the Server
// works: R730xd,R740xd
func (c *IloClient) GetServerPowerStateDell() (string, error) {
	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1"
	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return "", err
	}

	var data SystemViewDell

	json.Unmarshal(resp, &data)

	return data.PowerState, nil

}

//CheckLoginDell ... Will check the credentials of the Server
// works: R730xd,R740xd
func (c *IloClient) CheckLoginDell() (string, error) {
	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1"
	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return "", err
	}
	var data SystemViewDell
	json.Unmarshal(resp, &data)
	return string(data.Status.Health), nil
}

//CreateJobDell ... Create a Job based on the changed bios settings
/* Payload
   {"TargetSettingsURI":"/redfish/v1/Systems/System.Embedded.1/Bios/Settings"}
*/
func (c *IloClient) CreateJobDell(jsonData []byte) (string, error) {
	url := c.Hostname + "/redfish/v1/Managers/iDRAC.Embedded.1/Jobs"
	resp, _, err := queryData(c, "POST", url, jsonData)
	if err != nil {
		return "", err
	}
	var k JobResponseDell
	json.Unmarshal(resp, &k)
	return k.MessageExtendedInfo[0].Message, nil
}

func (c *IloClient) GetJobsStatusDell() ([]JobStatusDell, error) {
	url := c.Hostname + "/redfish/v1/Managers/iDRAC.Embedded.1/Jobs"
	var jobs []JobStatusDell
	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return jobs, err
	}
	var k MemberCountDell
	json.Unmarshal(resp, &k)
	for i := range k.Members {
		_url := c.Hostname + k.Members[i].OdataId
		resp, _, err := queryData(c, "GET", _url, nil)
		if err != nil {
			return jobs, err
		}
		var output JobStatusDell
		json.Unmarshal(resp, &output)
		jobs = append(jobs, output)
	}
	return jobs, nil
}

//SetBiosSettingsDell ... Set Bios Settings
/* Payload
{"Attributes":{"BootMode": "Bios"}}
*/
func (c *IloClient) SetBiosSettingsDell(jsonData []byte) (string, error) {
	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1/Bios/Settings"
	resp, _, err := queryData(c, "PATCH", url, jsonData)
	if err != nil {
		return "", err
	}
	var k JobResponseDell
	json.Unmarshal(resp, &k)
	return k.MessageExtendedInfo[0].Message, nil
}

//ClearJobsDell ... Deletes all the Jobs in the jobs queue
func (c *IloClient) ClearJobsDell() (string, error) {
	url := c.Hostname + "/redfish/v1/Managers/iDRAC.Embedded.1/Jobs"
	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return "", err
	}
	var k MemberCountDell
	json.Unmarshal(resp, &k)
	for i := range k.Members {
		_url := c.Hostname + k.Members[i].OdataId
		_, _, err := queryData(c, "DELETE", _url, nil)
		if err != nil {
			return "", err
		}
	}
	return "Jobs Deleted", nil
}

//SetAttributesDell ... Will set the Attributes for IDRAC,Lifecycle Attributes and System
/* Payload
{"Attributes":{"LCAttributes.1.AutoUpdate": "1"}}
*/
func (c *IloClient) SetAttributesDell(service string, jsonData []byte) (string, error) {
	var url string
	if service == "idrac" {
		url = c.Hostname + "/redfish/v1/Managers/iDRAC.Embedded.1/Attributes"
	} else if service == "lc" {
		url = c.Hostname + "/redfish/v1/Managers/LifecycleController.Embedded.1/Attributes"
	} else if service == "system" {
		url = c.Hostname + "/redfish/v1/Managers/System.Embedded.1/Attributes"
	}
	resp, _, err := queryData(c, "PATCH", url, jsonData)
	if err != nil {
		return "", err
	}
	var k JobResponseDell
	json.Unmarshal(resp, &k)
	return k.MessageExtendedInfo[0].Message, nil
}

//GetMacAddressDell ... Will fetch all the mac address of a particular Server
func (c *IloClient) GetMacAddressDell() (string, error) {
	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/"
	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return "", err
	}
	var x MemberCountDell
	var Macs []MACData
	json.Unmarshal(resp, &x)
	for i := range x.Members {
		_url := c.Hostname + x.Members[i].OdataId
		resp, _, err := queryData(c, "GET", _url, nil)
		if err != nil {
			return "", err
		}
		var y GetMacAddressDell
		json.Unmarshal(resp, &y)
		macData := MACData{
			Name:        y.ID,
			Description: y.Description,
			MacAddress:  y.MACAddress,
			Status:      y.Status.Health,
			State:       y.Status.State,
			Vlan:        y.VLAN,
		}
		Macs = append(Macs, macData)
	}
	output, _ := json.Marshal(Macs)
	return string(output), nil
}

//GetProcessorHealthDell ... Will Fetch the Processor Health Details
// works: R730xd,R740xd
func (c *IloClient) GetProcessorHealthDell() ([]HealthList, error) {
	///redfish/v1/Systems/System.Embedded.1/Processors

	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1/Processors"
	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var (
		x             ProcessorsListDataDell
		processHealth []HealthList
	)

	json.Unmarshal(resp, &x)

	for i := range x.Members {
		_url := c.Hostname + x.Members[i].OdataId
		resp, _, err := queryData(c, "GET", _url, nil)
		if err != nil {
			return nil, err
		}

		var y ProcessorDataDell

		json.Unmarshal(resp, &y)

		procHealth := HealthList{
			Name:   y.ID,
			Health: y.Status.Health,
			State:  y.Status.State,
		}
		processHealth = append(processHealth, procHealth)
	}

	return processHealth, nil

}

// func (c *IloClient) GetMemoryHealthDell() (string, error) {}

//GetPowerHealthDell ... Will Fetch the Power Health Details
// works: R730xd,R740xd
func (c *IloClient) GetPowerHealthDell() ([]HealthList, error) {
	url := c.Hostname + "/redfish/v1/Chassis/System.Embedded.1/Power"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var (
		x             PowerDataDell
		powerSupplies []HealthList
	)

	json.Unmarshal(resp, &x)

	if x.PowerSuppliescount != 0 {
		for i := range x.PowerSupplies {
			powerControlHealth := HealthList{
				Name:   x.PowerSupplies[i].MemberID,
				Health: x.PowerSupplies[i].Status.Health,
				State:  x.PowerSupplies[i].Status.State,
			}
			powerSupplies = append(powerSupplies, powerControlHealth)
		}
	}

	if x.Redundancycount != 0 {
		for i := range x.Redundancy {
			redundHealth := HealthList{
				Name:   x.Redundancy[i].Name,
				Health: x.Redundancy[i].Status.Health,
				State:  x.Redundancy[i].Status.State,
			}
			powerSupplies = append(powerSupplies, redundHealth)
		}
	}

	if x.Voltagescount != 0 {
		for i := range x.Voltages {
			voltageHealth := HealthList{
				Name:   x.Voltages[i].Name,
				Health: x.Voltages[i].Status.Health,
				State:  x.Voltages[i].Status.State,
			}
			powerSupplies = append(powerSupplies, voltageHealth)
		}
	}

	return powerSupplies, nil
}

//GetSensorsHealthDell ... Will Fetch the Sensors Health Details
// works: R730xd,R740xd
func (c *IloClient) GetSensorsHealthDell() ([]HealthList, error) {

	url := c.Hostname + "/redfish/v1/Chassis/System.Embedded.1/Thermal"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var (
		x             ThermalHealthListDell
		thermalHealth []HealthList
	)

	json.Unmarshal(resp, &x)

	// Fetching the Redundancy health info
	if x.Redundancycount != 0 {
		for i := range x.Redundancy {
			redundHealth := HealthList{
				Name:   x.Redundancy[i].Name,
				Health: x.Redundancy[i].Status.Health,
				State:  x.Redundancy[i].Status.State,
			}
			thermalHealth = append(thermalHealth, redundHealth)
		}
	}

	if x.Fanscount != 0 {
		for i := range x.Fans {
			fanHealth := HealthList{
				Name:   x.Fans[i].Name,
				Health: x.Fans[i].Status.Health,
				State:  x.Fans[i].Status.State,
			}
			thermalHealth = append(thermalHealth, fanHealth)
		}
	}

	if x.Temperaturescount != 0 {
		for i := range x.Temperatures {
			tempData := HealthList{
				Name:   x.Temperatures[i].Name,
				Health: x.Temperatures[i].Status.Health,
				State:  x.Temperatures[i].Status.State,
			}
			thermalHealth = append(thermalHealth, tempData)
		}
	}

	return thermalHealth, nil

}

//
func (c *IloClient) GetStorageDriveDetailsDell() ([]StorageDriveDetailsDell, error) {

	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1/Storage"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var (
		x          StorageCollectionDell
		_drivedata []StorageDriveDetailsDell
	)

	json.Unmarshal(resp, &x)

	for i := range x.Members {

		_url := c.Hostname + x.Members[i].OdataId
		resp, _, err := queryData(c, "GET", _url, nil)
		if err != nil {
			return nil, err
		}

		var y StorageDetailsDell

		json.Unmarshal(resp, &y)

		if y.Drivescount != 0 {
			for k := range y.Drives {
				_url := c.Hostname + y.Drives[k].OdataId
				resp, _, err := queryData(c, "GET", _url, nil)
				if err != nil {
					return nil, err
				}
				var z StorageDriveDetailsDell

				json.Unmarshal(resp, &z)

				_drivedata = append(_drivedata, z)
			}

		} else {
			continue
		}

	}
	return _drivedata, nil

}

//GetStorageHealthDell ... Will Fetch the Storage Health Details
// works: R730xd,R740xd
func (c *IloClient) GetStorageHealthDell() ([]StorageHealthList, error) {

	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1/Storage"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var (
		x           StorageCollectionDell
		_healthdata []StorageHealthList
	)

	json.Unmarshal(resp, &x)

	for i := range x.Members {

		_url := c.Hostname + x.Members[i].OdataId
		resp, _, err := queryData(c, "GET", _url, nil)
		if err != nil {
			return nil, err
		}

		var y StorageDetailsDell

		json.Unmarshal(resp, &y)

		storageHealth := StorageHealthList{
			Name:   y.ID,
			Health: y.Status.Health,
			State:  y.Status.State,
			Space:  0,
		}
		_healthdata = append(_healthdata, storageHealth)

		if y.Drivescount != 0 {
			for k := range y.Drives {
				_url := c.Hostname + y.Drives[k].OdataId
				resp, _, err := queryData(c, "GET", _url, nil)
				if err != nil {
					return nil, err
				}
				var z StorageDriveDetailsDell

				json.Unmarshal(resp, &z)

				storageHealth := StorageHealthList{
					Name:   z.Name,
					Health: z.Status.Health,
					State:  z.Status.State,
					Space:  z.CapacityBytes,
				}
				_healthdata = append(_healthdata, storageHealth)
			}

		} else {
			continue
		}

	}
	return _healthdata, nil

}

//GetAggHealthDataDell ... will fetch the data related to all components health(aggregated view)
func (c *IloClient) GetAggHealthDataDell(model string) ([]HealthList, error) {

	if strings.ToLower(model) == "r730xd" {

		return nil, nil

	} else if strings.ToLower(model) == "r740xd" {
		url := c.Hostname + "/redfish/v1/UpdateService/FirmwareInventory"

		resp, _, err := queryData(c, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		var (
			x           MemberCountDell
			_healthdata []HealthList
		)

		json.Unmarshal(resp, &x)

		for i := range x.Members {
			r, _ := regexp.Compile("Installed")
			if r.MatchString(x.Members[i].OdataId) == true {
				_url := c.Hostname + x.Members[i].OdataId
				resp, _, err := queryData(c, "GET", _url, nil)
				if err != nil {
					return nil, err
				}

				var y FirmwareDataDell

				json.Unmarshal(resp, &y)

				healthData := HealthList{
					Name:   y.Name,
					State:  y.Status.State,
					Health: y.Status.Health,
				}

				_healthdata = append(_healthdata, healthData)

			}
		}

		return _healthdata, nil
	}
	return nil, nil
}

//GetFirmwareDell ... will fetch the Firmware details
func (c *IloClient) GetFirmwareDell() ([]FirmwareData, error) {

	url := c.Hostname + "/redfish/v1/UpdateService/FirmwareInventory"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var (
		x         MemberCountDell
		_firmdata []FirmwareData
	)

	json.Unmarshal(resp, &x)

	for i := range x.Members {
		r, _ := regexp.Compile("Installed")
		if r.MatchString(x.Members[i].OdataId) == true {
			_url := c.Hostname + x.Members[i].OdataId
			resp, _, err := queryData(c, "GET", _url, nil)
			if err != nil {
				return nil, err
			}

			var y FirmwareDataDell

			json.Unmarshal(resp, &y)

			firmData := FirmwareData{
				Name:       y.Name,
				Id:         y.ID,
				Version:    y.Version,
				Updateable: y.Updateable,
			}

			_firmdata = append(_firmdata, firmData)

		}
	}

	return _firmdata, nil

}

//GetBiosDataDell ... will fetch the Bios Details
func (c *IloClient) GetBiosDataDell() (BiosAttributesData, error) {

	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1/Bios"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return BiosAttributesData{}, err
	}

	var x BiosAttrDell

	json.Unmarshal(resp, &x)

	return x.Attributes, nil

}

//GetLifecycleAttrDell ... will fetch the lifecycle attributes
func (c *IloClient) GetLifecycleAttrDell() (LifeCycleData, error) {

	url := c.Hostname + "/redfish/v1/Managers/LifecycleController.Embedded.1/Attributes"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return LifeCycleData{}, err
	}

	var x LifeCycleAttrDell

	json.Unmarshal(resp, &x)

	_data := x.Attributes

	_LfcycleDat := LifeCycleData{
		AutoBackup:                          _data.AutoBackup,
		AutoDiscovery:                       _data.AutoDiscovery,
		AutoUpdate:                          _data.AutoUpdate,
		BIOSRTDRequested:                    _data.BIOSRTDRequested,
		CollectSystemInventoryOnRestart:     _data.CollectSystemInventoryOnRestart,
		DiscoveryFactoryDefaults:            _data.DiscoveryFactoryDefaults,
		IPAddress:                           _data.IPAddress,
		IPChangeNotifyPS:                    _data.IPChangeNotifyPS,
		IgnoreCertWarning:                   _data.IgnoreCertWarning,
		Licensed:                            _data.Licensed,
		LifecycleControllerState:            _data.LifecycleControllerState,
		PartConfigurationUpdate:             _data.PartConfigurationUpdate,
		PartFirmwareUpdate:                  _data.PartFirmwareUpdate,
		ProvisioningServer:                  _data.ProvisioningServer,
		StorageHealthRollupStatus:           _data.StorageHealthRollupStatus,
		SystemID:                            _data.SystemID,
		UserProxyPassword:                   _data.UserProxyPassword,
		UserProxyPort:                       _data.UserProxyPort,
		UserProxyServer:                     _data.UserProxyServer,
		UserProxyType:                       _data.UserProxyType,
		UserProxyUserName:                   _data.UserProxyUserName,
		VirtualAddressManagementApplication: _data.VirtualAddressManagementApplication,
	}

	return _LfcycleDat, nil

}

//GetIDRACAttrDell ... will fetch the Idrac attributes
func (c *IloClient) GetIDRACAttrDell() (IDRACAttributesData, error) {

	url := c.Hostname + "/redfish/v1/Managers/iDRAC.Embedded.1/Attributes"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return IDRACAttributesData{}, err
	}

	var x IDRACAttrDell

	json.Unmarshal(resp, &x)

	return x.Attributes, nil

}

//GetSysAttrDell ... will fetch the System Attributes
func (c *IloClient) GetSysAttrDell() (SysAttributesData, error) {

	url := c.Hostname + "/redfish/v1/Managers/System.Embedded.1/Attributes"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return SysAttributesData{}, err
	}

	var x SysAttrDell

	json.Unmarshal(resp, &x)

	return x.Attributes, nil

}

//GetBootOrderDell ... will fetch the BootOrder Details
func (c *IloClient) GetBootOrderDell() ([]BootOrderData, error) {

	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1/BootSources"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var x BootOrderDell

	json.Unmarshal(resp, &x)

	var _bootOrder []BootOrderData

	for i := range x.Attributes.BootSeq {

		_result := BootOrderData{
			Enabled: x.Attributes.BootSeq[i].Enabled,
			Index:   x.Attributes.BootSeq[i].Index,
			Name:    x.Attributes.BootSeq[i].Name,
		}

		_bootOrder = append(_bootOrder, _result)
	}

	return _bootOrder, nil

}

//GetSystemEventLogsDell ... Fetch the System Event Logs from the Idrac
func (c *IloClient) GetSystemEventLogsDell(version string) ([]SystemEventLogRes, error) {

	url := c.Hostname + "/redfish/v1/Managers/iDRAC.Embedded.1/Logs/Sel"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	// v1, err := ver.NewVersion("3.15.17.15")
	v1, _ := ver.NewConstraint("<= 3.15.17.15")
	v2, _ := ver.NewConstraint("<= 3.21.26.22")
	v3, _ := ver.NewConstraint("> 3.21.26.22")
	v4, _ := ver.NewVersion(version)

	if v1.Check(v4) {

		var x SystemEventLogsV1Dell

		json.Unmarshal(resp, &x)

		var _systemEventLogs []SystemEventLogRes

		for i := range x.Members {

			_result := SystemEventLogRes{
				EntryCode:  x.Members[i].EntryCode[0].Member,
				Message:    x.Members[i].Message,
				Name:       x.Members[i].Name,
				SensorType: x.Members[i].SensorType[0].Member,
				Severity:   x.Members[i].Severity,
			}

			_systemEventLogs = append(_systemEventLogs, _result)
		}

		return _systemEventLogs, nil

	} else if v2.Check(v4) || v3.Check(v4) {

		var x SystemEventLogsV2Dell

		json.Unmarshal(resp, &x)

		var _systemEventLogs []SystemEventLogRes

		for i := range x.Members {

			_result := SystemEventLogRes{
				EntryCode:  x.Members[i].EntryCode,
				Message:    x.Members[i].Message,
				Name:       x.Members[i].Name,
				SensorType: x.Members[i].SensorType,
				Severity:   x.Members[i].Severity,
			}

			_systemEventLogs = append(_systemEventLogs, _result)
		}

		return _systemEventLogs, nil
	}
	return nil, err
}

//GetUserAccountsDell ... Fetch the current users created
func (c *IloClient) GetUserAccountsDell() ([]Accounts, error) {

	url := c.Hostname + "/redfish/v1/Managers/iDRAC.Embedded.1/Accounts"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var x MemberCountDell
	var users []Accounts

	json.Unmarshal(resp, &x)

	for i := range x.Members {
		_url := c.Hostname + x.Members[i].OdataId
		resp, _, err := queryData(c, "GET", _url, nil)
		if err != nil {
			return nil, err
		}

		var y AccountsInfoDell

		json.Unmarshal(resp, &y)

		user := Accounts{
			Name:     y.Name,
			Enabled:  y.Enabled,
			Locked:   y.Locked,
			RoleId:   y.RoleID,
			Username: y.UserName,
		}
		users = append(users, user)

	}

	return users, nil

}

//GetSystemInfoDell ... Will fetch the system info
func (c *IloClient) GetSystemInfoDell() (SystemData, error) {

	url := c.Hostname + "/redfish/v1/Systems/System.Embedded.1"

	resp, _, err := queryData(c, "GET", url, nil)
	if err != nil {
		return SystemData{}, err
	}

	var x SystemViewDell

	json.Unmarshal(resp, &x)

	_result := SystemData{Health: x.Status.Health,
		Memory:          x.MemorySummary.TotalSystemMemoryGiB,
		Model:           x.Model,
		PowerState:      x.PowerState,
		Processors:      x.ProcessorSummary.Count,
		ProcessorFamily: x.ProcessorSummary.Model,
		SerialNumber:    x.SerialNumber,
	}

	return _result, nil

}

//GetComponentAttr ... Will fetch all the component level attributes
//Supported values are: ALL, System, BIOS, IDRAC, NIC, FC, LifecycleController, RAID.
func (c *IloClient) GetComponentAttr(comp string) (ExportConfigResponse, error) {

	url := c.Hostname + "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ExportSystemConfiguration"
	data, _ := json.Marshal(map[string]interface{}{
		"ExportFormat": "JSON",
		"ShareParameters": map[string]interface{}{
			"Target": comp,
		},
	})

	_, header, err := queryData(c, "POST", url, []byte(data))
	if err != nil {
		return ExportConfigResponse{}, err
	}
	var taskURL string

	for k, v := range header {
		if k == "Location" {
			taskURL = v[0]
			break
		}
	}

	for {
		taskUrl := c.Hostname + taskURL

		resp, _, err := queryData(c, "GET", taskUrl, nil)
		if err != nil {
			return ExportConfigResponse{}, err
		}

		var x ExportConfigStatus

		json.Unmarshal(resp, &x)

		if x.TaskState == "Running" {
			x = ExportConfigStatus{}
			time.Sleep(time.Minute)
		} else {
			var y ExportConfigResponse
			json.Unmarshal(resp, &y)
			return y, nil
		}
	}

	return ExportConfigResponse{}, nil
}
