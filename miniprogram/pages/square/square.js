//square.js
const app = getApp();

Page({
  data: {
    //tabbar
    tabbar: {},
    TabCur: 0,
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
    ],
    list: [
      {
        id: 174,
        userid: 10,
        title: "日本岚山、和服一日游",
        banner: "https://hbimg.huabanimg.com/1ff95bdf3070e1fbff052a03ed353b409749f5ea16a809-WXy25b_fw658",
        points: 6,
        like: "62",
        userinfo: {
          id: 10,
          nickName: "李诗源",
          avatarurl: "https://pic3.zhimg.com/80/v2-fd0a58741fdf20f256c755719f81871e_hd.jpg"
        },
        islike: 0
      },
      {
        id: 173,
        userid: 9,
        title: "日本阿寒湖一日游",
        banner: "https://hbimg.huabanimg.com/ee5bf07b84fead3d57b445d2e7fa8eb6afe827c617e9c-ha1fZH_fw658",
        points: 7,
        like: "92",
        userinfo: {
          id: 9,
          nickName: "大飞狼",
          avatarurl: "https://pic3.zhimg.com/80/v2-fd0a58741fdf20f256c755719f81871e_hd.jpg"
        },
        islike: 0
      },
      {
        id: 172,
        userid: 8,
        title: "二次璧大乱斗东京动漫游",
        banner: "http://img1qn.moko.cc/2019-08-12/235e9bab-046e-4fea-afc2-4a049d81774e.jpg?imageView2/2/w/915/h/915/q/85",
        points: 4,
        like: "41",
        userinfo: {
          id: 8,
          nickName: "黄飞鸿",
          avatarurl: "https://pic3.zhimg.com/80/v2-fd0a58741fdf20f256c755719f81871e_hd.jpg"
        },
        islike: 0
      },
      {
        id: 100,
        userid: 314,
        title: "心和身体总要有一个在路上🏃",
        banner: "http://img.mb.moko.cc/2019-05-18/285bd040-2e62-4e1b-b0e8-91351c1f3c67.jpg?imageView2/2/w/915/h/915",
        points: 5,
        like: "110",
        userinfo: {
          id: 314,
          nickName: "二夏",
          avatarurl: "https://pic3.zhimg.com/80/v2-fd0a58741fdf20f256c755719f81871e_hd.jpg"
        },
        islike: 0
      },
      {
        id: 99,
        userid: 312,
        title: "新疆两日游",
        banner: "http://img.mb.moko.cc/2019-04-26/d4f1905c-3952-42be-9214-72260b97b0be.jpg?imageView2/2/w/915/h/915",
        points: 5,
        like: "99",
        userinfo: {
          id: 312,
          nickName: "Tohsaka",
          avatarurl: "https://pic3.zhimg.com/80/v2-fd0a58741fdf20f256c755719f81871e_hd.jpg"
        },
        islike: 0
      }
    ],
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
