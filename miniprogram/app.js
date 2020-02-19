//app.js
App({
  onLaunch: function () {
    //隐藏系统tabBar
    this.hideTabBar();
    this.getSystemInfo();

    if (!wx.cloud) {
      console.error('请使用 2.2.3 或以上的基础库以使用云能力')
    } else {
      wx.cloud.init({
        // env 参数说明：
        //   env 参数决定接下来小程序发起的云开发调用（wx.cloud.xxx）会默认请求到哪个云环境的资源
        //   此处请填入环境 ID, 环境 ID 可打开云控制台查看
        //   如不填则使用默认环境（第一个创建的环境）
        env: 'wuxie-test-id',
        traceUser: true,
      })
    }

    this.globalData = this.globalData;
  },
  hideTabBar() {
    wx.hideTabBar({
      fail: function () {
        setTimeout(function () { // 做了个延时重试一次，作为保底。
          wx.hideTabBar()
        }, 500)
      }
    });
  },
  onShow: function () {
    //隐藏系统tabbar
    wx.hideTabBar();
  },
  getSystemInfo: function () {
    let t = this;
    wx.getSystemInfo({
      success: e => {
        this.globalData.systemInfo = e;
        this.globalData.StatusBar = e.statusBarHeight;
        let capsule = wx.getMenuButtonBoundingClientRect();
        if (capsule) {
          this.globalData.Custom = capsule;
          this.globalData.CustomBar = capsule.bottom + capsule.top - e.statusBarHeight;
        } else {
          this.globalData.CustomBar = e.statusBarHeight + 50;
        }
      }
    })
  },
  editTabbar: function () {
    let tabbar = this.globalData.tabBar;
    let currentPages = getCurrentPages();
    let _this = currentPages[currentPages.length - 1];
    let pagePath = _this.route;
    (pagePath.indexOf('/') != 0) && (pagePath = '/' + pagePath);
    for (let i in tabbar.list) {
      tabbar.list[i].selected = false;
      (tabbar.list[i].pagePath == pagePath) && (tabbar.list[i].selected = true);
    }
    console.log(tabbar);
    _this.setData({
      tabbar: tabbar
    });
    console.log(_this);
  },
  globalData: {
    systemInfo: null,//客户端设备信息
    tabBar: {
      "backgroundColor": "#ffffff",
      "color": "#979795",
      "selectedColor": "#0081ff",
      "list": [
        {
          "pagePath": "/pages/index/index",
          "iconPath": "/images/shouye-gray.png",
          "selectedIconPath": "/images/shouye.png",
          "text": "首页"
        },
        {
          "pagePath": "/pages/square/square",
          "iconPath": "/images/all-gray.png",
          "selectedIconPath": "/images/all.png",
          "text": "广场"
        },
        {
          "pagePath": "/pages/databaseGuide/databaseGuide",
          "iconPath": "/images/zengjia.png",
          "selectedIconPath": "/images/zengjia.png",
          "isSpecial": true,
          "text": "打卡"
        },
        {
          "pagePath": "/pages/rank/rank",
          "iconPath": "/images/shangquan-gray.png",
          "selectedIconPath": "/images/shangquan.png",
          "text": "排位"
        },
        {
          "pagePath": "/pages/storageConsole/storageConsole",
          "iconPath": "/images/wodeguanzhu-gray.png",
          "selectedIconPath": "/images/wodeguanzhu.png",
          "text": "我的"
        }
      ],
      ColorList: [
        {
          title: '嫣红',
          name: 'red',
          color: '#e54d42'
        },
        {
          title: '桔橙',
          name: 'orange',
          color: '#f37b1d'
        },
        {
          title: '明黄',
          name: 'yellow',
          color: '#fbbd08'
        },
        {
          title: '橄榄',
          name: 'olive',
          color: '#8dc63f'
        },
        {
          title: '森绿',
          name: 'green',
          color: '#39b54a'
        },
        {
          title: '天青',
          name: 'cyan',
          color: '#1cbbb4'
        },
        {
          title: '海蓝',
          name: 'blue',
          color: '#0081ff'
        },
        {
          title: '姹紫',
          name: 'purple',
          color: '#6739b6'
        },
        {
          title: '木槿',
          name: 'mauve',
          color: '#9c26b0'
        },
        {
          title: '桃粉',
          name: 'pink',
          color: '#e03997'
        },
        {
          title: '棕褐',
          name: 'brown',
          color: '#a5673f'
        },
        {
          title: '玄灰',
          name: 'grey',
          color: '#8799a3'
        },
        {
          title: '草灰',
          name: 'gray',
          color: '#aaaaaa'
        },
        {
          title: '墨黑',
          name: 'black',
          color: '#333333'
        },
        {
          title: '雅白',
          name: 'white',
          color: '#ffffff'
        },
      ]
    }
  }
})
