## **About GoCache-Driver**
gocache-driver是gocache的一个驱动，主要负责简化go程序与gocache的交互，从而提高你对gocache相关套件的开发效率。
* 主要功能及语句

    创建cachedb
    ```go
    db:=New(hostname,port,password,database)*CacheDB
    ```
    连接cachedb
    ```go
    db.Connect()error    //error != nil 就代表连接失败了
    ```
    添加或修改节点
    ```go
    db.SetKey(key,value string)error
    ```
    查找节点
    ```go
    db.GetKeys(keys ...string)[]string,error
    db.GetAllKeys()map[string][]byte,error //获取整个数据库全部键值对
    ```
    删除节点
    ```go
    db.DeleteKeys(keys ...string)error
    ```
    创建数据库
    ```go
    db.CreateDB()error
    ```
    删除数据库
    ```go
    db.DropDB()error
    ```
    手动储存
    ```go
    db.Save()error
    ```