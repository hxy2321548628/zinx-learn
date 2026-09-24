package znet

import (
	"fmt"
	"strconv"
	"zinx-learn/internal/zinx/zitface"
)

type MsgHandler struct {
	Apis map[uint32]zitface.IRouter //存放每个MsgId 所对应的处理方法的map属性
}

func (mh *MsgHandler) DoMsgHandler(request zitface.IRequest) {
	router, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), " is not FOUND!")
		return
	}

	router.PreHandle(request)
	router.Handle(request)
	router.PostHandle(request)

}
func (mh *MsgHandler) AddRouter(msgId uint32, router zitface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	//2 添加msg与api的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}

func NewMsgHandler() zitface.IMsgHandle {
	return &MsgHandler{
		Apis: make(map[uint32]zitface.IRouter),
	}
}
