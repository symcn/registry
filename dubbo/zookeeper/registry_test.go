/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package zookeeper

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/symcn/registry/dubbo/common"
	"github.com/symcn/registry/dubbo/common/constant"
	"github.com/symcn/registry/dubbo/remoting/zookeeper"
)

func Test_Register(t *testing.T) {
	regurl, _ := common.NewURL("registry://127.0.0.1:1111", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER)))
	url, _ := common.NewURL("dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider", common.WithParamsValue(constant.CLUSTER_KEY, "mock"), common.WithParamsValue("serviceid", "soa.mock"), common.WithMethods([]string{"GetUser", "AddUser"}))

	ts, reg, _ := newMockZkRegistry(&regurl)
	defer ts.Stop()
	err := reg.Register(&url)
	children, _ := reg.client.GetChildren("/dubbo/com.ikurento.user.UserProvider/providers")
	assert.Regexp(t, ".*dubbo%3A%2F%2F127.0.0.1%3A20000%2Fcom.ikurento.user.UserProvider%3Fanyhost%3Dtrue%26cluster%3Dmock%26.*.serviceid%3Dsoa.mock", children)
	assert.NoError(t, err)
}

func Test_UnRegister(t *testing.T) {
	// register
	regurl, _ := common.NewURL("registry://127.0.0.1:1111", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER)))
	url, _ := common.NewURL("dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider", common.WithParamsValue(constant.CLUSTER_KEY, "mock"), common.WithParamsValue("serviceid", "soa.mock"), common.WithMethods([]string{"GetUser", "AddUser"}))

	ts, reg, _ := newMockZkRegistry(&regurl)
	defer ts.Stop()
	err := reg.Register(&url)
	children, _ := reg.client.GetChildren("/dubbo/com.ikurento.user.UserProvider/providers")
	assert.Regexp(t, ".*dubbo%3A%2F%2F127.0.0.1%3A20000%2Fcom.ikurento.user.UserProvider%3Fanyhost%3Dtrue%26cluster%3Dmock%26.*.serviceid%3Dsoa.mock", children)
	assert.NoError(t, err)

	err = reg.UnRegister(&url)
	children, err = reg.client.GetChildren("/dubbo/com.ikurento.user.UserProvider/providers")
	assert.Equal(t, 0, len(children))
	assert.Error(t, err)
	assert.True(t, reg.IsAvailable())

	err = reg.Register(&url)
	children, _ = reg.client.GetChildren("/dubbo/com.ikurento.user.UserProvider/providers")
	assert.Regexp(t, ".*dubbo%3A%2F%2F127.0.0.1%3A20000%2Fcom.ikurento.user.UserProvider%3Fanyhost%3Dtrue%26cluster%3Dmock%26.*.serviceid%3Dsoa.mock", children)
	assert.NoError(t, err)

}

func Test_Subscribe(t *testing.T) {
	regurl, _ := common.NewURL("registry://127.0.0.1:1111", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER)))
	url, _ := common.NewURL("dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider", common.WithParamsValue(constant.CLUSTER_KEY, "mock"), common.WithMethods([]string{"GetUser", "AddUser"}))
	ts, reg, _ := newMockZkRegistry(&regurl)

	//provider register
	err := reg.Register(&url)
	assert.NoError(t, err)

	if err != nil {
		return
	}

	//consumer register
	regurl.SetParam(constant.ROLE_KEY, strconv.Itoa(common.CONSUMER))
	_, reg2, _ := newMockZkRegistry(&regurl, zookeeper.WithTestCluster(ts))

	reg2.Register(&url)
	listener, _ := reg2.DoSubscribe(&url)

	serviceEvent, _ := listener.Next()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Regexp(t, ".*ServiceEvent{Action{add}.*", serviceEvent.String())
	defer ts.Stop()
}

func Test_UnSubscribe(t *testing.T) {
	regurl, _ := common.NewURL("registry://127.0.0.1:1111", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER)))
	url, _ := common.NewURL("dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider", common.WithParamsValue(constant.CLUSTER_KEY, "mock"), common.WithMethods([]string{"GetUser", "AddUser"}))
	ts, reg, _ := newMockZkRegistry(&regurl)

	//provider register
	err := reg.Register(&url)
	assert.NoError(t, err)

	if err != nil {
		return
	}

	//consumer register
	regurl.SetParam(constant.ROLE_KEY, strconv.Itoa(common.CONSUMER))
	_, reg2, _ := newMockZkRegistry(&regurl, zookeeper.WithTestCluster(ts))

	reg2.Register(&url)
	listener, _ := reg2.DoSubscribe(&url)

	serviceEvent, _ := listener.Next()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Regexp(t, ".*ServiceEvent{Action{add}.*", serviceEvent.String())

	reg2.UnSubscribe(&url, nil)
	assert.Nil(t, reg2.listener)

	defer ts.Stop()
}

func Test_ConsumerDestory(t *testing.T) {
	regurl, _ := common.NewURL("registry://127.0.0.1:1111", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.CONSUMER)))
	url, _ := common.NewURL("dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider", common.WithParamsValue(constant.CLUSTER_KEY, "mock"), common.WithMethods([]string{"GetUser", "AddUser"}))

	ts, reg, err := newMockZkRegistry(&regurl)
	defer ts.Stop()

	assert.NoError(t, err)
	err = reg.Register(&url)
	assert.NoError(t, err)
	_, err = reg.DoSubscribe(&url)
	assert.NoError(t, err)

	//listener.Close()
	time.Sleep(1e9)
	reg.Destroy()
	assert.Equal(t, false, reg.IsAvailable())

}

func Test_ProviderDestory(t *testing.T) {
	regurl, _ := common.NewURL("registry://127.0.0.1:1111", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER)))
	url, _ := common.NewURL("dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider", common.WithParamsValue(constant.CLUSTER_KEY, "mock"), common.WithMethods([]string{"GetUser", "AddUser"}))

	ts, reg, err := newMockZkRegistry(&regurl)
	defer ts.Stop()

	assert.NoError(t, err)
	reg.Register(&url)

	//listener.Close()
	time.Sleep(1e9)
	reg.Destroy()
	assert.Equal(t, false, reg.IsAvailable())
}
