## sunmary
Because the kubernetes event can only be saved for one hour by default in etcd.
Sometimes we need to find events to locate some reasons. But the event has been deleted by kubernetes apiserver.

So,we need a way to save the kubernetes event. this project watch the kubernetes all event. dump all event to mongo db.
and offer some interface support user access.


## Run
./kevents --kubeconfig=k8s1-/Users/shaxiangyu/conf/config --mongo-address 127.0.0.1 --mongo-db=admin --mongo-user mongoadmin  --mongo-passwd secret


## Note
因为 k8s event 的数据比较多，将 event 持久化到 mongo 中，我们需要对定期清理 mongo 中的event数据。 我们默认是只保存1天的数据。

具体的策略是每小时执行一次清理的任务， 清理任务会删除比当前时间小于等于1天的 event.