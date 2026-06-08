import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { AuditLog } from '../models';

@Component({
  selector: 'app-audit',
  template: `
    <div>
      <h2 class="bank-title">审计日志</h2>

      <div class="bank-card">
        <div class="flex-between mb-3">
          <div class="flex gap-2">
            <select class="bank-input" style="width: 150px;" [(ngModel)]="module" (change)="loadLogs()">
              <option value="">全部模块</option>
              <option value="auth">认证</option>
              <option value="account">账户</option>
              <option value="transfer">转账</option>
              <option value="api">API</option>
            </select>
            <input type="text" class="bank-input" style="width: 200px;" placeholder="搜索操作..." [(ngModel)]="keyword" (keyup.enter)="loadLogs()" />
          </div>
          <div style="color: #666; font-size: 14px;">
            共 {{ total }} 条记录 | 日志保留7年
          </div>
        </div>

        <table class="bank-table">
          <thead>
            <tr>
              <th>时间</th>
              <th>模块</th>
              <th>操作</th>
              <th>描述</th>
              <th>IP地址</th>
            </tr>
          </thead>
          <tbody>
            <tr *ngFor="let log of logs">
              <td>{{ log.created_at | date:'yyyy-MM-dd HH:mm:ss' }}</td>
              <td>
                <span class="bank-badge bank-badge-info">{{ getModuleName(log.module) }}</span>
              </td>
              <td style="font-weight: 500;">{{ log.action }}</td>
              <td>{{ log.description }}</td>
              <td style="color: #999; font-size: 13px;">{{ log.ip_address || '-' }}</td>
            </tr>
          </tbody>
        </table>

        <div class="flex-between mt-3" *ngIf="total > 0">
          <button class="bank-btn bank-btn-outline" (click)="prevPage()" [disabled]="page <= 1">上一页</button>
          <span>第 {{ page }} 页</span>
          <button class="bank-btn bank-btn-outline" (click)="nextPage()">下一页</button>
        </div>
      </div>

      <div class="bank-card mt-3" style="background: #fff8e1;">
        <h4 style="color: #8d6e00; margin-bottom: 12px;">📌 关于审计日志</h4>
        <ul style="color: #666; font-size: 13px; padding-left: 20px;">
          <li class="mb-1">所有操作均会记录审计日志，包括登录、转账、修改密码等敏感操作</li>
          <li class="mb-1">审计日志采用 HMAC 链式校验，确保日志不可篡改</li>
          <li class="mb-1">审计日志保留期限为 7 年，符合金融监管要求</li>
          <li>审计日志仅允许查询，不允许修改和删除</li>
        </ul>
      </div>
    </div>
  `
})
export class AuditComponent implements OnInit {
  logs: AuditLog[] = [];
  total = 0;
  page = 1;
  pageSize = 20;
  module = '';
  keyword = '';

  private apiUrl = 'http://localhost:8080/api';

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.loadLogs();
  }

  loadLogs(): void {
    let url = `${this.apiUrl}/audit/logs?page=${this.page}&page_size=${this.pageSize}`;
    if (this.module) {
      url += `&module=${this.module}`;
    }
    if (this.keyword) {
      url += `&action=${this.keyword}`;
    }

    this.http.get<any>(url).subscribe({
      next: (data) => {
        this.logs = data.logs || [];
        this.total = data.total || 0;
      }
    });
  }

  prevPage(): void {
    if (this.page > 1) {
      this.page--;
      this.loadLogs();
    }
  }

  nextPage(): void {
    this.page++;
    this.loadLogs();
  }

  getModuleName(mod: string): string {
    const map: Record<string, string> = {
      'auth': '认证',
      'account': '账户',
      'transfer': '转账',
      'audit': '审计',
      'api': 'API'
    };
    return map[mod] || mod;
  }
}
