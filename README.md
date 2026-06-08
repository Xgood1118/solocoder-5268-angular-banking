# Angular + Go 全栈个人银行系统

一个用于学习和教学的银行核心系统 Demo，完整实现了银行业务核心概念：账户余额计算、转账幂等、对账一致性、敏感数据保护、审计日志、限额风控、双因子认证等。

## 技术栈

### 后端
- **语言**: Go 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: PostgreSQL
- **缓存/计数器**: Redis
- **认证**: JWT + bcrypt + pepper
- **定时任务**: robfig/cron

### 前端
- **框架**: Angular 17
- **UI组件**: Angular Material
- **样式**: 自定义银行风格 (深蓝 + 金色)
- **状态管理**: RxJS + BehaviorSubject

## 核心功能

### 账户管理
- 开户、销户、冻结/解冻
- 支持活期、定期、支票账户
- 多币种支持 (CNY/USD/EUR)
- **借贷记账法**: 每笔交易写借贷流水，余额 = SUM(credit) - SUM(debit)

### 转账系统
- 同行转账 / 跨行转账 (Mock 清算接口)
- 实时到账 / 普通到账
- **幂等控制**: 客户端带 biz_id (UUID)，服务端去重
- **状态机**: pending → frozen → processing → success/failed
- **并发控制**: 乐观锁 + 事务
- **余额非负**: 校验余额充足

### 对账引擎
- 每日凌晨 0 点跑批对账 (系统账 vs 流水账)
- 差异自动识别 (金额不符、缺流水、多流水)
- 人工对账工单
- 对账差异报告

### 审计日志
- 所有操作写审计流水
- **不可篡改**: HMAC 链式校验
- 日志保留 7 年
- 支持按模块、操作、时间查询

### 限额风控
- 单笔限额、单日限额、月度限额
- **Redis 计数器**: INCR + EXPIRE
- 超限实时拒绝

### 双因子认证
- 登录、修改密码、绑卡需验证码
- 6 位数字、5 分钟有效
- 3 次错误锁定
- 不能连续两次相同

### 敏感数据脱敏
- 身份证、手机号、卡号显示脱敏 (仅后4位)
- 日志自动脱敏

### 利率计算
- 日利率 = 年利率 / 360 (行业惯例)
- 按日计息，月末结息
- 活期/定期不同利率

## 项目结构

```
5268-angular-banking/
├── backend/                    # Go 后端
│   ├── cmd/server/             # 应用入口
│   ├── internal/
│   │   ├── account/            # 账户服务
│   │   ├── transfer/           # 转账服务
│   │   ├── audit/              # 审计日志
│   │   ├── limit/              # 限额风控
│   │   ├── recon/              # 对账引擎
│   │   ├── auth/               # 认证 & 双因子
│   │   ├── interest/           # 利率计算
│   │   └── report/             # 报表
│   └── pkg/
│       ├── database/           # 数据库
│       ├── redis/              # Redis
│       └── masking/            # 脱敏工具
├── frontend/                   # Angular 前端
│   └── src/app/
│       ├── core/               # 核心服务 & 组件
│       ├── login/              # 登录注册
│       ├── account/            # 账户管理
│       ├── transfer/           # 转账 & 流水
│       ├── recon/              # 对账中心
│       ├── audit/              # 审计日志
│       └── settings/           # 设置
└── docker-compose.yml          # PostgreSQL + Redis
```

## 快速开始

### 1. 启动基础设施

```bash
docker-compose up -d
```

### 2. 启动后端

```bash
cd backend
go mod download
go run cmd/server/main.go
```

后端默认运行在 `http://localhost:8080`

### 3. 启动前端

```bash
cd frontend
npm install
npm start
```

前端默认运行在 `http://localhost:4200`

## API 接口

### 认证
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/twofa/verify` - 双因子验证

### 账户
- `GET /api/accounts` - 账户列表
- `POST /api/accounts` - 开立账户
- `GET /api/accounts/:id` - 账户详情
- `POST /api/accounts/:id/freeze` - 冻结账户
- `POST /api/accounts/:id/unfreeze` - 解冻账户
- `DELETE /api/accounts/:id` - 销户
- `GET /api/accounts/:id/ledger` - 账户流水

### 转账
- `POST /api/transfers` - 创建转账 (幂等)
- `GET /api/transfers` - 转账记录列表
- `GET /api/transfers/:id` - 转账详情
- `GET /api/transfers/biz/:biz_id` - 按业务号查询

### 限额
- `GET /api/limits` - 查询限额
- `POST /api/limits` - 设置限额

### 审计
- `GET /api/audit/logs` - 审计日志列表

### 对账
- `GET /api/recon/reports` - 对账报告
- `GET /api/recon/reports/:id/differences` - 对账差异
- `POST /api/recon/trigger` - 触发对账

## 设计要点

### 余额计算
使用借贷记账法，不直接 UPDATE 余额字段，而是通过流水计算。每笔交易生成两条流水（一借一贷），余额 = SUM(credit) - SUM(debit)。

### 转账幂等
客户端生成唯一 biz_id (UUID)，服务端用唯一索引去重。重复请求直接返回之前的结果，不会重复扣款。

### 并发安全
使用乐观锁 (version 字段) + 数据库事务，避免 ABA 问题。

### 跨行转账状态机
```
pending → frozen → processing → success
                            ↘ failed → rolled_back
```

1. 冻结金额
2. 发送清算请求 (异步)
3. 收到清算回执
4. 确认成功/回滚

### 对账机制
- 系统账: accounts.balance
- 流水账: SUM(ledger_entries)
- 每日对比，差异生成报告
- 支持人工对账

### 审计日志不可篡改
每条日志记录前一条日志的 HMAC，形成链式结构。篡改任何一条都会导致后续 HMAC 不匹配。

### 限额校验
使用 Redis INCR + EXPIRE 实现计数器，比数据库聚合快得多，且天然支持过期。

## 学习要点

这个项目涵盖了金融系统设计中常见的坑：

1. **余额计算**: 为什么用借贷记账法而不是直接 update？
2. **幂等性**: 网络重试时如何保证不重复扣款？
3. **并发控制**: 同时转两笔，余额怎么算？
4. **一致性**: 系统账和流水账对不上怎么办？
5. **安全**: 敏感数据怎么保护？密码怎么存？
6. **审计**: 操作日志怎么防篡改？
7. **风控**: 限额怎么实现才高效？
8. **双因子**: 验证码设计的最佳实践

## 注意事项

这是一个教学 Demo，请勿用于生产环境。实际银行系统需要：
- 真正的硬件加密机 (HSM)
- 符合监管的安全审计
- 真正的支付清算接口
- 多活部署和容灾
- 完整的合规体系

## License

MIT
