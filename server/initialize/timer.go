package initialize

import (
	"fmt"
	"github.com/flipped-aurora/gin-vue-admin/server/service/content"
	"github.com/flipped-aurora/gin-vue-admin/server/task"

	"github.com/robfig/cron/v3"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

func Timer() {
	go func() {
		var option []cron.Option
		option = append(option, cron.WithSeconds())
		// 清理DB定时任务
		_, err := global.GVA_Timer.AddTaskByFunc("ClearDB", "@daily", func() {
			err := task.ClearTable(global.GVA_DB) // 定时任务方法定在task文件包中
			if err != nil {
				fmt.Println("timer error:", err)
			}
		}, "定时清理数据库【日志，黑名单】内容", option...)
		if err != nil {
			fmt.Println("add timer error:", err)
		}

		// 办公工具临时目录：默认 24 小时后删除
		_, err = global.GVA_Timer.AddTaskByFunc("CleanOfficeToolsTemp", "0 0 */1 * * *", func() {
			if n, cerr := content.CleanOfficeToolsExpired(); cerr != nil {
				global.GVA_LOG.Warn("办公工具临时文件清理失败", zap.Error(cerr))
			} else if n > 0 {
				global.GVA_LOG.Debug("办公工具定时清理", zap.Int("removed", n))
			}
		}, "清理办公工具超过保留期的临时文件", option...)
		if err != nil {
			fmt.Println("add office tools cleanup timer error:", err)
		}

		// 其他定时任务定在这里 参考上方使用方法

		//_, err := global.GVA_Timer.AddTaskByFunc("定时任务标识", "corn表达式", func() {
		//	具体执行内容...
		//  ......
		//}, option...)
		//if err != nil {
		//	fmt.Println("add timer error:", err)
		//}
	}()
}
