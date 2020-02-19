const app = getApp();
Component({
  /**
   * 组件的属性列表
   */
  properties: {
    tabbar: {
      type: Object,
      value: {
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
        ]
      }
    }
  },

  /**
   * 组件的初始数据
   */
  data: {

  },
  /**
   * 组件的方法列表
   */
  methods: {

  }
})
