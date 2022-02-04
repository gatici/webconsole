// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
//

/*
 * Connectivity Service Configuration
 *
 * APIs to configure connectivity service in Aether Network
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package configapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omec-project/webconsole/backend/logger"
	"github.com/omec-project/webconsole/configmodels"
)

// DeviceGroupGroupNameDelete -
func DeviceGroupGroupNameDelete(c *gin.Context) {
	logger.ConfigLog.Debugf("DeviceGroupGroupNameDelete")
	if ret := DeviceGroupDeleteHandler(c); ret == true {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
}

// DeviceGroupGroupNamePut -
func DeviceGroupGroupNamePut(c *gin.Context) {
	logger.ConfigLog.Debugf("DeviceGroupGroupNamePut")
	if ret := DeviceGroupPostHandler(c, configmodels.Put_op); ret == true {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
}

// DeviceGroupGroupNamePatch -
func DeviceGroupGroupNamePatch(c *gin.Context) {
	logger.ConfigLog.Debugf("DeviceGroupGroupNamePatch")
	c.JSON(http.StatusOK, gin.H{})
}

// DeviceGroupGroupNamePost -
func DeviceGroupGroupNamePost(c *gin.Context) {
	logger.ConfigLog.Debugf("DeviceGroupGroupNamePost")
	if ret := DeviceGroupPostHandler(c, configmodels.Post_op); ret == true {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
}

// NetworkSliceSliceNameDelete -
func NetworkSliceSliceNameDelete(c *gin.Context) {
	logger.ConfigLog.Debugf("Received NetworkSliceSliceNameDelete ")
	if ret := NetworkSliceDeleteHandler(c); ret == true {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
}

// NetworkSliceSliceNamePost -
func NetworkSliceSliceNamePost(c *gin.Context) {
	logger.ConfigLog.Debugf("Received NetworkSliceSliceNamePost ")
	if ret := NetworkSlicePostHandler(c, configmodels.Post_op); ret == true {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
}

// NetworkSliceSliceNamePut -
func NetworkSliceSliceNamePut(c *gin.Context) {
	logger.ConfigLog.Debugf("Received NetworkSliceSliceNamePut ")
	if ret := NetworkSlicePostHandler(c, configmodels.Put_op); ret == true {
		c.JSON(http.StatusOK, gin.H{})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
}
