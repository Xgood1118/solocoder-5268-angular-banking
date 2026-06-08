import { Component, OnInit } from '@angular/core';
import { AuthService } from '../core/auth.service';
import { User, LimitInfo } from '../models';
import { HttpClient } from '@angular/common/http';
import { MatDialog } from '@angular/material/dialog';
import { TwoFADialogComponent } from '../core/twofa-dialog.component';
import { ConfirmDialogComponent } from '../core/confirm-dialog.component';

@Component({
  selector: 'app-settings',
  template: `
    <div>
      <h2 class="bank-title">账户设置</h2>

      <div class="grid-2">
        <div class="bank-card">
          <h3 class="bank-subtitle">个人信息</h3>
          <div *ngIf="user" style="font-size: 14px;">
            <div class="flex-between mb-2" style="padding-bottom: 12px; border-bottom: 1px solid #eee;">
              <span style="color: #666;">用户名</span>
              <span>{{ user.username }}</span>
            </div>
            <div class="flex-between mb-2" style="padding-bottom: 12px; border-bottom: 1px solid #eee;">
              <span style="color: #666;">姓名</span>
              <span>{{ user.full_name }}</span>
            </div>
            <div class="flex-between mb-2" style="padding-bottom: 12px; border-bottom: 1px solid #eee;">
              <span style="color: #666;">手机号</span>
              <span>{{ user.phone }}</span>
            </div>
            <div class="flex-between mb-2" style="padding-bottom: 12px; border-bottom: 1px solid #eee;">
              <span style="color: #666;">邮箱</span>
              <span>{{ user.email }}</span>
            </div>
            <div class="flex-between">
              <span style="color: #666;">身份证号</span>
              <span>{{ user.id_card }}</span>
            </div>
          </div>
        </div>

        <div class="bank-card">
          <h3 class="bank-subtitle">安全设置</h3>

          <div class="mb-3">
            <div class="flex-between" style="margin-bottom: 8px;">
              <span>登录密码</span>
              <button class="bank-btn bank-btn-outline" style="padding: 6px 16px; font-size: 13px;" (click)="showPasswordModal = true">
                修改密码
              </button>
            </div>
            <p style="color: #999; font-size: 13px;">定期修改密码可提高账户安全性</p>
          </div>

          <div class="mb-3">
            <div class="flex-between" style="margin-bottom: 8px;">
              <span>双因子认证</span>
              <span class="bank-badge" [ngClass]="user?.twofa_enabled ? 'bank-badge-success' : 'bank-badge-warning'">
                {{ user?.twofa_enabled ? '已启用' : '未启用' }}
              </span>
            </div>
            <p style="color: #999; font-size: 13px;">启用后登录时需要短信验证码</p>
          </div>

          <div>
            <div class="flex-between" style="margin-bottom: 8px;">
              <span>交易限额</span>
            </div>
            <p style="color: #999; font-size: 13px;">单笔: ¥50,000 | 单日: ¥200,000 | 单月: ¥500,000</p>
          </div>
        </div>

        <div class="bank-card">
          <h3 class="bank-subtitle">限额设置</h3>
          <div *ngIf="limits" style="font-size: 14px;">
            <div class="flex-between mb-2" style="padding-bottom: 12px; border-bottom: 1px solid #eee;">
              <span style="color: #666;">单笔限额</span>
              <span>¥{{ limits.per_transaction.limit | number:'1.2-2' }}</span>
            </div>
            <div class="flex-between mb-2" style="padding-bottom: 12px; border-bottom: 1px solid #eee;">
              <span style="color: #666;">日累计限额</span>
              <span>¥{{ limits.daily.limit | number:'1.2-2' }}</span>
            </div>
            <div class="flex-between">
              <span style="color: #666;">月累计限额</span>
              <span>¥{{ limits.monthly.limit | number:'1.2-2' }}</span>
            </div>
          </div>
        </div>

        <div class="bank-card">
          <h3 class="bank-subtitle">安全提示</h3>
          <ul style="color: #666; font-size: 13px; padding-left: 20px;">
            <li class="mb-2">请勿向任何人透露您的密码和验证码</li>
            <li class="mb-2">银行工作人员不会向您索要密码</li>
            <li class="mb-2">请定期检查账户交易记录</li>
            <li class="mb-2">如遇可疑情况请立即联系客服</li>
            <li>客服热线：400-888-8888</li>
          </ul>
        </div>
      </div>

      <div *ngIf="showPasswordModal" class="bank-modal-overlay" (click.self)="showPasswordModal = false">
        <div class="bank-modal">
          <div class="bank-modal-header">
            <span>修改密码</span>
            <button class="bank-modal-close" (click)="showPasswordModal = false">&times;</button>
          </div>
          <div *ngIf="passwordError" class="alert alert-error">{{ passwordError }}</div>
          <div *ngIf="passwordSuccess" class="alert alert-success">密码修改成功！</div>
          <form (ngSubmit)="changePassword()">
            <div class="bank-form-group">
              <label class="bank-label">原密码</label>
              <input type="password" class="bank-input" [(ngModel)]="passwordForm.oldPassword" name="oldPassword" required />
            </div>
            <div class="bank-form-group">
              <label class="bank-label">新密码</label>
              <input type="password" class="bank-input" [(ngModel)]="passwordForm.newPassword" name="newPassword" required minlength="8" />
              <small style="color: #999;">至少8位字符</small>
            </div>
            <div class="bank-form-group">
              <label class="bank-label">确认新密码</label>
              <input type="password" class="bank-input" [(ngModel)]="passwordForm.confirmPassword" name="confirmPassword" required />
            </div>
            <div class="bank-modal-footer">
              <button type="button" class="bank-btn bank-btn-outline" (click)="showPasswordModal = false">取消</button>
              <button type="submit" class="bank-btn bank-btn-gold" [disabled]="passwordLoading">
                {{ passwordLoading ? '提交中...' : '确认修改' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  `
})
export class SettingsComponent implements OnInit {
  user: User | null = null;
  limits: any = null;
  showPasswordModal = false;
  passwordForm = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  };
  passwordError = '';
  passwordSuccess = false;
  passwordLoading = false;

  private apiUrl = 'http://localhost:8080/api';

  constructor(
    private authService: AuthService,
    private http: HttpClient,
    private dialog: MatDialog
  ) {}

  ngOnInit(): void {
    this.user = this.authService.getCurrentUser();
    this.loadLimits();
  }

  loadLimits(): void {
    this.http.get(`${this.apiUrl}/limits?scope=transfer`).subscribe({
      next: (data) => {
        this.limits = data;
      }
    });
  }

  changePassword(): void {
    if (this.passwordForm.newPassword !== this.passwordForm.confirmPassword) {
      this.passwordError = '两次输入的新密码不一致';
      return;
    }

    if (this.passwordForm.newPassword.length < 8) {
      this.passwordError = '新密码至少需要8位字符';
      return;
    }

    this.passwordLoading = true;
    this.passwordError = '';
    this.passwordSuccess = false;

    this.authService.changePassword(this.passwordForm.oldPassword, this.passwordForm.newPassword).subscribe({
      next: () => {
        this.passwordLoading = false;
        this.passwordSuccess = true;
        setTimeout(() => {
          this.showPasswordModal = false;
          this.passwordSuccess = false;
          this.passwordForm = { oldPassword: '', newPassword: '', confirmPassword: '' };
        }, 2000);
      },
      error: (err) => {
        this.passwordLoading = false;
        this.passwordError = err.error?.error || '密码修改失败';
      }
    });
  }
}
