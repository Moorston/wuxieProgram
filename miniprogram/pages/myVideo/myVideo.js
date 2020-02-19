//index.js
const app = getApp();


Page({
  data: {
    //tabbar
    tabbar: {}
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
