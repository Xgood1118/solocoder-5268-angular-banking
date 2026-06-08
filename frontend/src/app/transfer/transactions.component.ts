import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Transfer } from '../models';

@Component({
  selector: 'app-transactions',
  template: `
    <div>
      <h2 class="bank-title">交易明细</h2>

      <div class="bank-card">
        <div class="flex-between mb-3">
          <div class="flex gap-2">
            <select class="bank-input" style="width: 150px;" [(ngModel)]="status" (change)="loadTransfers()">
              <option value="">全部状态</option>
              <option value="success">成功</option>
              <option value="pending">处理中</option>
              <option value="failed">失败</option>
            </select>
          </div>
          <div style="color: #666; font-size: 14px;">
            共 {{ total }} 条记录
          </div>
        </div>

        <table class="bank-table">
          <thead>
            <tr>
              <th>时间</th>
              <th>类型</th>
              <th>收款人</th>
              <th>金额</th>
              <th>手续费</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr *ngFor="let t of transfers">
              <td>{{ t.created_at | date:'yyyy-MM-dd HH:mm:ss' }}</td>
              <td>
                <span class="bank-badge" [ngClass]="getTypeClass(t.transfer_type)">
                  {{ t.transfer_type === 'intra_bank' ? '同行' : '跨行' }}
                </span>
              </td>
              <td>
                <div>{{ t.to_account_name }}</div>
                <div style="font-size: 12px; color: #999;">{{ t.to_account_no }}</div>
              </td>
              <td class="amount-negative">-¥{{ t.amount | number:'1.2-2' }}</td>
              <td>¥{{ t.fee | number:'1.2-2' }}</td>
              <td>
                <span class="bank-badge" [ngClass]="getStatusClass(t.status)">
                  {{ getStatusName(t.status) }}
                </span>
              </td>
              <td>
                <button class="bank-btn bank-btn-outline" style="padding: 4px 12px; font-size: 12px;" (click)="viewDetail(t)">详情</button>
              </td>
            </tr>
            <tr *ngIf="transfers.length === 0">
              <td colspan="7" style="text-align: center; padding: 40px; color: #999;">
                暂无交易记录
              </td>
            </tr>
          </tbody>
        </table>

        <div class="flex-between mt-3" *ngIf="total > 0">
          <button class="bank-btn bank-btn-outline" (click)="prevPage()" [disabled]="page <= 1">上一页</button>
          <span>第 {{ page }} 页 / 共 {{ totalPages }} 页</span>
          <button class="bank-btn bank-btn-outline" (click)="nextPage()" [disabled]="page >= totalPages">下一页</button>
        </div>
      </div>

      <div *ngIf="selectedTransfer" class="bank-modal-overlay" (click.self)="selectedTransfer = null">
        <div class="bank-modal">
          <div class="bank-modal-header">
            <span>交易详情</span>
            <button class="bank-modal-close" (click)="selectedTransfer = null">&times;</button>
          </div>
          <div style="font-size: 14px;">
            <div class="flex-between mb-1"><span style="color: #666;">交易编号</span><span>{{ selectedTransfer.biz_id }}</span></div>
            <div class="flex-between mb-1"><span style="color: #666;">交易类型</span><span>{{ selectedTransfer.transfer_type === 'intra_bank' ? '同行转账' : '跨行转账' }}</span></div>
            <div class="flex-between mb-1"><span style="color: #666;">到账方式</span><span>{{ selectedTransfer.transfer_speed === 'realtime' ? '实时到账' : '普通到账' }}</span></div>
            <div class="flex-between mb-1"><span style="color: #666;">转出账户</span><span>{{ selectedTransfer.from_account_no }}</span></div>
            <div class="flex-between mb-1"><span style="color: #666;">收款人</span><span>{{ selectedTransfer.to_account_name }}</span></div>
            <div class="flex-between mb-1"><span style="color: #666;">收款账号</span><span>{{ selectedTransfer.to_account_no }}</span></div>
            <div *ngIf="selectedTransfer.to_bank_name" class="flex-between mb-1"><span style="color: #666;">收款银行</span><span>{{ selectedTransfer.to_bank_name }}</span></div>
            <div class="flex-between mb-1"><span style="color: #666;">转账金额</span><span style="font-size: 20px; font-weight: 700; color: var(--danger);">-¥{{ selectedTransfer.amount | number:'1.2-2' }}</span></div>
            <div class="flex-between mb-1"><span style="color: #666;">手续费</span><span>¥{{ selectedTransfer.fee | number:'1.2-2' }}</span></div>
            <div class="flex-between mb-1"><span style="color: #666;">状态</span><span class="bank-badge" [ngClass]="getStatusClass(selectedTransfer.status)">{{ getStatusName(selectedTransfer.status) }}</span></div>
            <div class="flex-between mb-1"><span style="color: #666;">提交时间</span><span>{{ selectedTransfer.created_at | date:'yyyy-MM-dd HH:mm:ss' }}</span></div>
            <div *ngIf="selectedTransfer.completed_at" class="flex-between mb-1"><span style="color: #666;">完成时间</span><span>{{ selectedTransfer.completed_at | date:'yyyy-MM-dd HH:mm:ss' }}</span></div>
            <div *ngIf="selectedTransfer.failure_reason" class="flex-between"><span style="color: #666;">失败原因</span><span style="color: var(--danger);">{{ selectedTransfer.failure_reason }}</span></div>
            <div class="mt-2" *ngIf="selectedTransfer.description">
              <div style="color: #666; margin-bottom: 4px;">附言</div>
              <div>{{ selectedTransfer.description }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  `
})
export class TransactionsComponent implements OnInit {
  transfers: Transfer[] = [];
  total = 0;
  page = 1;
  pageSize = 20;
  status = '';
  selectedTransfer: Transfer | null = null;

  private apiUrl = 'http://localhost:8080/api';

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.loadTransfers();
  }

  loadTransfers(): void {
    let url = `${this.apiUrl}/transfers?page=${this.page}&page_size=${this.pageSize}`;
    if (this.status) {
      url += `&status=${this.status}`;
    }

    this.http.get<any>(url).subscribe({
      next: (data) => {
        this.transfers = data.transfers || [];
        this.total = data.total || 0;
      }
    });
  }

  prevPage(): void {
    if (this.page > 1) {
      this.page--;
      this.loadTransfers();
    }
  }

  get totalPages(): number {
    return Math.ceil(this.total / this.pageSize);
  }

  nextPage(): void {
    if (this.page < this.totalPages) {
      this.page++;
      this.loadTransfers();
    }
  }

  viewDetail(t: Transfer): void {
    this.selectedTransfer = t;
  }

  getStatusName(status: string): string {
    const map: Record<string, string> = {
      'success': '成功',
      'pending': '处理中',
      'frozen': '已冻结',
      'processing': '清算中',
      'failed': '失败',
      'rolled_back': '已回滚'
    };
    return map[status] || status;
  }

  getStatusClass(status: string): string {
    const map: Record<string, string> = {
      'success': 'bank-badge-success',
      'pending': 'bank-badge-warning',
      'frozen': 'bank-badge-warning',
      'processing': 'bank-badge-info',
      'failed': 'bank-badge-danger',
      'rolled_back': 'bank-badge-danger'
    };
    return map[status] || 'bank-badge-info';
  }

  getTypeClass(type: string): string {
    return type === 'intra_bank' ? 'bank-badge-info' : 'bank-badge-warning';
  }
}
