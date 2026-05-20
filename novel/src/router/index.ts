import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import FullScreenLayout from '@/layouts/FullScreenLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: MainLayout,
      redirect: '/books',
      children: [
        {
          path: 'books',
          name: 'Books',
          component: () => import('@/views/Books.vue'),
          meta: { title: '书本管理' }
        },
        {
          path: 'prompts',
          name: 'Prompts',
          component: () => import('@/views/Prompts.vue'),
          meta: { title: '提示词管理' }
        },
        {
          path: 'prompt-preview',
          name: 'PromptPreview',
          component: () => import('@/views/PromptPreview.vue'),
          meta: { title: '提示词预览' }
        },
        {
          path: 'creative',
          name: 'Creative',
          component: () => import('@/views/Creative.vue'),
          meta: { title: '抽卡区' }
        },
        {
          path: 'workflow',
          name: 'Workflow',
          component: () => import('@/views/Workflow.vue'),
          meta: { title: '工作流' }
        },
        {
          path: 'analysis/:bookId',
          name: 'BookAnalysis',
          component: () => import('@/views/BookAnalysis.vue'),
          meta: { title: '拆书库' }
        },
        {
          path: 'characters/:bookId',
          name: 'CharacterLibrary',
          component: () => import('@/views/CharacterLibrary.vue'),
          meta: { title: '角色库' }
        },
        {
          path: 'config',
          name: 'Config',
          component: () => import('@/views/Config.vue'),
          meta: { title: 'API配置' }
        },
        {
          path: 'profile',
          name: 'Profile',
          component: () => import('@/views/Profile.vue'),
          meta: { title: '个人中心' }
        },
        {
          path: 'experience-shares',
          name: 'ExperienceShares',
          component: () => import('@/views/ExperienceShares.vue'),
          meta: { title: '经验分享' }
        }
      ]
    },
    {
      path: '/write/:bookId',
      component: FullScreenLayout,
      children: [
        {
          path: '',
          name: 'Write',
          component: () => import('@/views/Write.vue'),
          meta: { title: '创作' }
        }
      ]
    },
    {
      path: '/workflow/:id',
      component: FullScreenLayout,
      children: [
        {
          path: '',
          name: 'WorkflowWorkbench',
          component: () => import('@/views/WorkflowWorkbench.vue'),
          meta: { title: '工作流工作台' }
        }
      ]
    },
    {
      path: '/experience-shares/:id/edit',
      component: FullScreenLayout,
      children: [
        {
          path: '',
          name: 'ExperienceShareEditor',
          component: () => import('@/views/ExperienceShareEditor.vue'),
          meta: { title: '编辑经验卡片' }
        }
      ]
    },
    {
      path: '/experience-shares/:id',
      component: FullScreenLayout,
      children: [
        {
          path: '',
          name: 'ExperienceShareDetail',
          component: () => import('@/views/ExperienceShareDetail.vue'),
          meta: { title: '经验卡片详情' }
        }
      ]
    },
    {
      path: '/book-analysis-prompts',
      component: FullScreenLayout,
      children: [
        {
          path: '',
          name: 'BookAnalysisPrompts',
          component: () => import('@/views/BookAnalysisPrompts.vue'),
          meta: { title: '拆书库提示词管理' }
        }
      ]
    }
  ],
})

export default router
