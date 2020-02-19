//rank.js
const app = getApp();

Page({
  data: {
    //tabbar
    tabbar: {},
    CustomBar: app.globalData.CustomBar,
    option1: [
      { text: '全部', value: 0 },
      { text: '今日', value: 1 },
      { text: '今周', value: 2 }
    ],
    value1: 0,
  },

  onLoad: function () {
    app.hideTabBar();
    app.editTabbar();
    if (!wx.cloud) {
      wx.redirectTo({
        url: '../chooseLib/chooseLib',
      })
      return
    }
  },
})