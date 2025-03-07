package rabbitmq

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

/*
对 RabbitMQ 函数库的再次封装，便于操作
*/

type RabbitMQ struct {
	channel  *amqp.Channel
	Name     string
	exchange string
}

/*创建新结构体*/
func New(s string) *RabbitMQ {
	conn, e := amqp.Dial(s)
	if e != nil {
		log.Fatalln(e)
	}

	ch, e := conn.Channel()
	if e != nil {
		log.Fatalln(e)
	}

	q, e := ch.QueueDeclare(
		"",    // 名字
		false, // 耐用
		true,  // 不使用时删除
		false, // 专用的
		false, // 不等待
		nil,   // 参数
	)

	if e != nil {
		log.Fatalln(e)
	}

	// 分配内存，构建一个新结构体
	mq := new(RabbitMQ)
	mq.channel = ch
	mq.Name = q.Name

	// 返回对象指针
	return mq
}

/*将 RabbitMQMQ 结构体的消息队列与一个 exchange 绑定*/
func (q *RabbitMQ) Bind(exchange string) {
	e := q.channel.QueueBind(q.Name, "", exchange, false, nil)
	if e != nil {
		log.Fatalln(e)
	}
	q.exchange = exchange
}

/*往某个消息队列发送消息*/
func (q *RabbitMQ) Send(queue string, body interface{}) {
	// 序列化
	str, e := json.Marshal(body)
	if e != nil {
		log.Fatalln(e)
	}
	e = q.channel.Publish("", queue, false, false, amqp.Publishing{
		ReplyTo: q.Name,
		Body:    []byte(str),
	})
	if e != nil {
		log.Fatalln(e)
	}
}

/*往某个 exchange 发送消息*/
func (q *RabbitMQ) Publish(exchange string, body interface{}) {
	// 序列化
	str, e := json.Marshal(body)
	if e != nil {
		log.Fatalln(e)
	}
	e = q.channel.Publish(exchange, "", false, false, amqp.Publishing{
		ReplyTo: q.Name,
		Body:    []byte(str),
	})
	if e != nil {
		log.Fatalln(e)
	}
}

/*生成一个接收信息的管道*/
func (q *RabbitMQ) Consume() <-chan amqp.Delivery {
	c, e := q.channel.Consume(q.Name, "", true, false, false, false, nil)
	if e != nil {
		log.Fatalln(e)
	}
	return c
}

/*关闭消息队列*/
func (q *RabbitMQ) Close() {
	q.channel.Close()
}
