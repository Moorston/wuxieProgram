//index.js
const app = getApp()

Page({
  data: {
    //判断小程序的API，回调，参数，组件等是否在当前版本可用。
    canIUse: wx.canIUse('button.open-type.getUserInfo')
  },

  //需要在onLoad函数里判断用户是否被授权
  onLoad: function (options) {
    var that = this;
    // 查看是否授权
    wx.getSetting({
      success: function (res) {
        if (res.authSetting['scope.userInfo']) {
          //用户已经授权过则跳转到小程序首页，注意这里登陆界面会出现一闪而过的现象
          wx.switchTab({
            url: '/pages/index/index'
          })
        }
      }
    })
  },

  //button组件的bindgetuserinfo绑定的函数
  bindGetUserInfo: function (e) {
    if (e.detail.userInfo) {
      //用户按了允许授权按钮
      var that = this;

      that.login(); //声明的函数 

      // 查询数据库是否存在当前用户，避免同一用户被添加两次
      const db = wx.cloud.database();
      db.collection('user').where({
        _openid: app.globalData.openid
      }).get({
        success: res => {
          if (res.data.length == 0) { //若数据库无该用户的信息则添加到数据库
            const db = wx.cloud.database()
            db.collection('user').add({
              data: {
                openid: app.globalData.openid,
                nickName: e.detail.userInfo.nickName,
                avatarUrl: e.detail.userInfo.avatarUrl,
                province: e.detail.userInfo.province,
                city: e.detail.userInfo.city,
                groupid:0
              },
              header: {
                'content-type': 'application/json'
              },
              success: function (res) {
                console.log("插入小程序登录用户信息成功！");
              }
            })
          }
        },
      })

      //授权成功后，跳转进入小程序首页
      let url = '/pages/index/index'
      wx.switchTab({
        url: url,
        success(res) {
          let page = getCurrentPages().pop();
          if (page == undefined || page == null) {
            return
          }
          page.onLoad();
        }
      })
    } else {
      //用户按了拒绝按钮
      wx.showModal({
        title: '警告',
        content: '您点击了拒绝授权，将无法进入小程序，请授权之后再进入!!!',
        showCancel: false,
        confirmText: '返回授权',
        success: function (res) {
          if (res.confirm) {
            console.log('用户点击了“返回授权”')
          }
        }
      })
    }
  },
  //调用名为‘login’的云函数获取用户接口，这个如果选用小程序云开发，初始会给login云函数
  login() {
    let that = this;
    wx.cloud.callFunction({
      name: 'login',
      complete: res => {
        console.log('云函数获取到的openid: ', res.result.openid)
        app.globalData.openId = res.result.openid; //给全局变量openId和userInfo赋值
        app.globalData.userInfo = res.result.event.userInfo;
      }
    })
  },

})
