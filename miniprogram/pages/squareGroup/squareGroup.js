//square.js
const app = getApp();

Page({
  data: {
    //tabbar
    tabbar: {},
    TabCur: 1,
    CustomBar: app.globalData.CustomBar,
    scrollLeft: 0,

    tabList: [
      {
        id: 0,
        name: "广场",
        url: ""
      },
      {
        id: 1,
        name: "考核组",
        url: ""
      },
    ]
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

  tabSelect(e) {
    this.setData({
      TabCur: e.currentTarget.dataset.id,
      scrollLeft: (e.currentTarget.dataset.id - 1) * 60
    })
  }

})
