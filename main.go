package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/ying32/govcl/pkgs/winappres"
	"github.com/ying32/govcl/vcl"
	"github.com/ying32/govcl/vcl/types"
	"github.com/ying32/govcl/vcl/types/colors"
	"github.com/ying32/govcl/vcl/win"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 获取配置文件
type Configs map[string]json.RawMessage

var configPath = "./config.json"

type Desc struct {
	Time    string `json:"time"`
	Title   string `json:"title"`
	Message string `json:"message"`
}
type Location struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type MainConfig struct {
	Data     []Desc   `json:"data"`
	Location Location `json:"location"`
}

type NowWeather struct {
	Code       string `json:"code"`
	UpdateTime string `json:"updateTime"`
	FxLink     string `json:"fxLink"`
	Now        struct {
		ObsTime   string `json:"obsTime"`
		Temp      string `json:"temp"`
		FeelsLike string `json:"feelsLike"`
		Icon      string `json:"icon"`
		Text      string `json:"text"`
		Wind360   string `json:"wind360"`
		WindDir   string `json:"windDir"`
		WindScale string `json:"windScale"`
		WindSpeed string `json:"windSpeed"`
		Humidity  string `json:"humidity"`
		Precip    string `json:"precip"`
		Pressure  string `json:"pressure"`
		Vis       string `json:"vis"`
		Cloud     string `json:"cloud"`
		Dew       string `json:"dew"`
	} `json:"now"`
	Refer struct {
		Sources []string `json:"sources"`
		License []string `json:"license"`
	} `json:"refer"`
}

type Weather7d struct {
	Code       string `json:"code"`
	UpdateTime string `json:"updateTime"`
	FxLink     string `json:"fxLink"`
	Daily      []struct {
		FxDate         string `json:"fxDate"`
		Sunrise        string `json:"sunrise"`
		Sunset         string `json:"sunset"`
		Moonrise       string `json:"moonrise"`
		Moonset        string `json:"moonset"`
		MoonPhase      string `json:"moonPhase"`
		TempMax        string `json:"tempMax"`
		TempMin        string `json:"tempMin"`
		IconDay        string `json:"iconDay"`
		TextDay        string `json:"textDay"`
		IconNight      string `json:"iconNight"`
		TextNight      string `json:"textNight"`
		Wind360Day     string `json:"wind360Day"`
		WindDirDay     string `json:"windDirDay"`
		WindScaleDay   string `json:"windScaleDay"`
		WindSpeedDay   string `json:"windSpeedDay"`
		Wind360Night   string `json:"wind360Night"`
		WindDirNight   string `json:"windDirNight"`
		WindScaleNight string `json:"windScaleNight"`
		WindSpeedNight string `json:"windSpeedNight"`
		Humidity       string `json:"humidity"`
		Precip         string `json:"precip"`
		Pressure       string `json:"pressure"`
		Vis            string `json:"vis"`
		Cloud          string `json:"cloud"`
		UvIndex        string `json:"uvIndex"`
	} `json:"daily"`
	Refer struct {
		Sources []string `json:"sources"`
		License []string `json:"license"`
	} `json:"refer"`
}

var conf *MainConfig
var confs Configs

var instanceOnce sync.Once

//从配置文件中载入json字符串
func LoadConfig(path string) (Configs, *MainConfig) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panicln("load config conf failed: ", err)
	}
	mainConfig := &MainConfig{}
	err = json.Unmarshal(buf, mainConfig)
	if err != nil {
		log.Panicln("decode config file failed:", string(buf), err)
	}
	allConfigs := make(Configs, 0)
	err = json.Unmarshal(buf, &allConfigs)
	if err != nil {
		log.Panicln("decode config file failed:", string(buf), err)
	}
	return allConfigs, mainConfig
}

func Init(path string) *MainConfig {
	if conf != nil && path != configPath {
		log.Printf("the config is already initialized, oldPath=%s, path=%s", configPath, path)
	}
	instanceOnce.Do(func() {
		allConfigs, mainConfig := LoadConfig(path)
		configPath = path
		conf = mainConfig
		confs = allConfigs
	})

	return conf
}

var closeFlag = false
var day = time.Now().Day()

type TMainForm struct {
	*vcl.TForm
	Text        *vcl.TLabel
	Timer1      *vcl.TTimer
	button      *vcl.TButton
	TTrayIcon   *vcl.TTrayIcon
	city        *vcl.TLabel
	weatherIcon *vcl.TImage
	listView    *vcl.TListView
	listColumn  *vcl.TListColumn
	tPanel      *vcl.TPanel
}

type TForm1 struct {
	*vcl.TForm
	Text *vcl.TLabel
}

var (
	mainForm *TMainForm
	form1    *TForm1
)

// 文件数据
var data []Desc
var locationData Location
var nowWeather NowWeather
var weather7d Weather7d

func main() {
	path := configPath
	//fmt.Println("path: ", path)
	Init(path)
	value := confs["data"]
	location := confs["location"]
	err3 := json.Unmarshal(location, &locationData)
	if err3 != nil {

	}
	//fmt.Println(string(value))

	// 和风key
	key := "0df1e18e68604b879dfa7911971e642e"

	res, _ := http.Get("https://devapi.qweather.com/v7/weather/now?key=" + key + "&location=" + locationData.ID)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	err1 := json.Unmarshal(body, &nowWeather)
	//fmt.Println(nowWeather.Now.Text)
	if err1 != nil {
		log.Panicln("天气failed:", "x", err1)
	}
	// 7天天气
	res7d, _ := http.Get("https://devapi.qweather.com/v7/weather/7d?key=" + key + "&location=" + locationData.ID)
	defer res7d.Body.Close()
	body7d, _ := ioutil.ReadAll(res7d.Body)

	err7d := json.Unmarshal(body7d, &weather7d)
	//fmt.Println(weather7d.Daily)
	if err7d != nil {
		log.Panicln("7天天气failed:", "x", err7d)
	}

	err := json.Unmarshal(value, &data)
	//fmt.Println(strings.Split(data[0].Time, ":")[0])
	//fmt.Println(len(data))
	if err != nil {
		log.Panicln("decode config file failed:", "x", err)
	}
	// 打开是否显示主窗口
	vcl.Application.SetShowMainForm(false)
	vcl.Application.SetOnMinimize(func(sender vcl.IObject) {
		mainForm.Hide() // 主窗口最隐藏掉
	})

	pm := vcl.NewPopupMenu(mainForm)
	item := vcl.NewMenuItem(mainForm)
	item.SetCaption("显示(&S)")
	item.SetOnClick(func(vcl.IObject) {
		mainForm.Show()
		vcl.Application.Restore()
		// Windows上为了最前面显示，有时候要调用SetForegroundWindow
		win.SetForegroundWindow(mainForm.Handle())
	})
	pm.Items().Add(item)

	item = vcl.NewMenuItem(mainForm)
	item.SetCaption("退出(&E)")
	item.SetOnClick(func(vcl.IObject) {
		// 主窗口关闭
		mainForm.Close()
		// 或者使用
		//		vcl.Application.Terminate()
	})
	pm.Items().Add(item)
	vcl.Application.SetMainFormOnTaskBar(true)

	trayIcon := vcl.NewTrayIcon(mainForm)
	trayIcon.SetVisible(true)
	trayIcon.SetPopupMenu(pm)
	trayIcon.SetOnClick(func(vcl.IObject) {
		//f.SetVisible(true)
		trayIcon.SetBalloonTitle("test")
		trayIcon.SetBalloonTimeout(10000)
		trayIcon.SetBalloonHint("我是提示正文啦")
		trayIcon.ShowBalloonHint()
		vcl.Application.SetShowMainForm(true)
		if mainForm.CanFocus() {
			mainForm.Hide() // 主窗口最隐藏掉
		} else {
			mainForm.Show()
			vcl.Application.Restore()
		}
		//win.SetForegroundWindow(mainForm.Handle())
		// 前置窗口
		//mainForm.BringToFront()
	})

	vcl.RunApp(&mainForm, &form1)
}

// --------------MainForm -----------------

func (f *TMainForm) click(object vcl.IObject) {
	form1.Show()
}

func (f *TMainForm) OnFormCreate(sender vcl.IObject) {

	f.SetCaption("打卡提示")
	f.SetColor(colors.ClWhite)
	// 标题栏不显示
	//f.EnabledSystemMenu(false)
	// 设置边框, 0 会禁止拉伸和移动
	f.SetBorderStyle(1)

	// 禁止最大化
	f.EnabledMaximize(false)

	f.SetWidth(400)
	f.SetHeight(600)
	f.ScreenCenter()

	//f.TTrayIcon.SetIcon(vcl.Application.Icon()) //不设置会自动使用Application.Icon
	//fmt.Println(vcl.Screen.WorkAreaHeight())

	//f.TMonthCalendar = vcl.NewMonthCalendar(f)
	//f.TMonthCalendar.SetParent(f)
	//f.TMonthCalendar.SetTop(100)

	// 时间
	timer := time.Now()
	sMonth := fmt.Sprintf("%02d", int(timer.Month()))
	sDay := fmt.Sprintf("%02d", timer.Day())
	sHour := fmt.Sprintf("%02d", timer.Hour())
	sMinute := fmt.Sprintf("%02d", timer.Minute())
	sSecond := fmt.Sprintf("%02d", timer.Second())
	f.Text = vcl.NewLabel(f)
	f.Text.SetParent(f)
	//f.Text.SetWidth(400)
	f.Text.SetHeight(30)
	//f.Text.SetColor(colors.ClBlack)
	f.Text.SetAlignment(types.AkLeft)
	f.Text.SetLayout(types.AkRight)
	f.Text.SetCaption(strconv.Itoa(timer.Year()) + "-" + sMonth + "-" + sDay + " " + sHour + ":" + sMinute + ":" + sSecond)
	f.Text.Font().SetSize(14)
	f.Text.Font().SetColor(colors.ClBlue)
	f.Text.SetWordWrap(true)
	// 居中
	f.Text.AnchorHorizontalCenterTo(f)
	//f.Text.AnchorVerticalCenterTo(f)
	f.Timer1 = vcl.NewTimer(f)
	f.Timer1.SetInterval(1000)
	f.Timer1.SetOnTimer(f.doTimer)

	// 实况天气
	//城市名称
	city := vcl.NewLabel(f)
	city.SetCaption(locationData.Name)
	city.SetParent(f)
	city.SetHeight(40)
	city.AnchorHorizontalCenterTo(f)
	city.SetTop(f.Text.Height())
	city.SetLeft(0)
	city.Font().SetSize(16)

	// 天气图标
	weatherIcon := vcl.NewImage(f)
	weatherIcon.SetParent(f)
	weatherIcon.Picture().LoadFromFile("./icons/" + nowWeather.Now.Icon + ".png")
	weatherIcon.SetHeight(64)
	weatherIcon.SetWidth(64)
	weatherIcon.SetTop(70)
	weatherIcon.AnchorHorizontalCenterTo(f)
	// 天气状态 + 温度
	weatherText := vcl.NewLabel(f)
	weatherText.AnchorHorizontalCenterTo(f)
	weatherText.SetCaption(nowWeather.Now.Text + "  " + nowWeather.Now.Temp + "℃")
	weatherText.SetParent(f)
	weatherText.SetTop(70 + 64)

	// 7天天气
	for i, v := range weather7d.Daily {
		top := int32((i+1)*30) + 70 + 104
		dayTxt := vcl.NewLabel(f)
		if i == 0 {
			dayTxt.SetCaption("今日")
		} else {
			dayTxt.SetCaption(v.FxDate[5:len(v.FxDate)])
		}

		dayTxt.SetWidth(100)
		dayTxt.SetTop(top)
		dayTxt.SetLeft(10)
		dayTxt.SetParent(f)

		dayIcon := vcl.NewLabel(f)
		dayIcon.SetCaption(v.TextDay)
		dayIcon.SetTop(top)
		dayIcon.SetLeft(100)
		dayIcon.SetParent(f)
		dayIcon.AnchorHorizontalCenterTo(f)

		dayTemp := vcl.NewLabel(f)
		dayTemp.SetCaption(v.TempMin + "~" + v.TempMax + "℃")
		//dayTemp.SetWidth(150)
		dayTemp.SetTop(top)
		//fmt.Println(dayTemp.Width())
		dayTemp.SetLeft(400 - dayTemp.Width())
		//dayTemp.SetColor()
		dayTemp.SetAlignment(types.TaRightJustify)
		dayTemp.SetParent(f)

		//dayTemp.SetAlign(types.AlClient)
		//dayTemp.SetParent(mainForm)
		//dayTemp.SetAutoSize(false)
		//dayTemp.Font().SetSize(13)
		dayTemp.SetAlignment(types.TaCenter)
		dayTemp.SetLayout(types.TlCenter)
	}

	btn := vcl.NewButton(f)
	btn.SetCaption("open")
	btn.SetParent(f)
	btn.SetOnClick(func(sender vcl.IObject) {
		form1.Show()
	})
	//weatherTemp := vcl.NewLabel(f)
	//weatherTemp.AnchorHorizontalCenterTo(f)
	//weatherTemp.SetCaption(nowWeather.Now.Temp + "℃")
	//weatherTemp.SetParent(f)
	//weatherTemp.SetTop(70 + 64 + 30)
	//f.button = vcl.NewButton(f)
	//f.button.SetParent(f)
	//f.button.SetCaption("点击")
	//f.button.SetOnClick(func(sender vcl.IObject) {
	//	form1.Show()
	//})
}

// 定时器
func (f *TMainForm) doTimer(sender vcl.IObject) {
	vcl.ThreadSync(func() {
		timer := time.Now()
		if day != timer.Day() {
			day = timer.Day()
			closeFlag = false
		}
		sMonth := fmt.Sprintf("%02d", int(timer.Month()))
		sDay := fmt.Sprintf("%02d", timer.Day())
		sHour := fmt.Sprintf("%02d", timer.Hour())
		hour := strconv.Itoa(timer.Hour())
		minute := strconv.Itoa(timer.Minute())
		sMinute := fmt.Sprintf("%02d", timer.Minute())
		sSecond := fmt.Sprintf("%02d", timer.Second())
		mainForm.Text.SetCaption(strconv.Itoa(timer.Year()) + "-" + sMonth + "-" + sDay + " " + sHour + ":" + sMinute + ":" + sSecond)

		for j := 0; j < len(data); j++ {
			dataTime := strings.Split(data[j].Time, ":")
			if hour == dataTime[0] {
				if minute == dataTime[1] && !closeFlag && !form1.CanFocus() {
					form1.Show()
					// 真 置顶
					form1.SetFormStyle(types.FsSystemStayOnTop)
					form1.Text.SetCaption(data[j].Message)
					// 显示到最前面
					form1.BringToFront()
					win.SetForegroundWindow(mainForm.Handle())
				} else if minute != dataTime[1] && closeFlag {
					closeFlag = false
				}
			}
		}
	})
}

func (f *TMainForm) OnFormCloseQuery(Sender vcl.IObject, CanClose *bool) {
	*CanClose = vcl.MessageDlg("是否退出？", types.MtConfirmation, types.MbYes, types.MbNo) == types.IdYes
}

func (f *TMainForm) OnButton1Click(object vcl.IObject) {
	form1.Show()
	//form1.ScreenCenter()
	//fmt.Println()
}

// ---------- Form1 ----------------
// 提示窗口
func (f *TForm1) OnFormCreate(sender vcl.IObject) {
	f.SetCaption("")
	f.SetWidth(200)
	f.SetHeight(300)
	f.SetLeft(vcl.Screen.WorkAreaWidth() - 200 - 10)
	f.SetTop(vcl.Screen.WorkAreaHeight() - 300 - 30)
	f.EnabledMinimize(false)
	f.EnabledMaximize(false)
	f.SetColor(colors.ClWhite)

	f.Text = vcl.NewLabel(f)
	f.Text.SetParent(f)
	f.Text.SetWidth(400)
	f.Text.SetHeight(400)
	//f.Text.SetColor(colors.ClBlack)
	f.Text.SetAlignment(types.AsrCenter)
	f.Text.SetLayout(types.AkRight)
	f.Text.SetCaption("")
	f.Text.Font().SetSize(14)
	f.Text.Font().SetColor(colors.ClBlack)
	f.Text.SetWordWrap(true)
	// 居中
	f.Text.AnchorHorizontalCenterTo(f)
	f.Text.AnchorVerticalCenterTo(f)

	form1.SetOnClose(func(sender vcl.IObject, action *types.TCloseAction) {
		form1.Hide()
		closeFlag = true
	})

}

func (f *TForm1) OnButton1Click(object vcl.IObject) {
	vcl.ShowMessage("Click")
}
