import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import FullScreenLayout from '@/layouts/FullScreenLayout.vue'
import TutorialLayout from '@/layouts/TutorialLayout.vue'

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
          path: 'prompts/pack/:name',
          name: 'PackDetail',
          component: () => import('@/views/PackDetail.vue'),
          meta: { title: '卡包详情' }
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
          path: 'ai-ranking',
          name: 'AiRanking',
          component: () => import('@/views/AiRanking.vue'),
          meta: { title: 'AI扫榜' }
        },
        {
          path: 'agent',
          name: 'Agent',
          component: () => import('@/views/Agent.vue'),
          meta: { title: '星芒 Agent' }
        },
        {
          path: 'wiki-graph',
          name: 'WikiGraph',
          component: () => import('@/views/WikiGraph.vue'),
          meta: { title: 'wiki图谱' }
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
          path: 'plugins',
          name: 'Plugins',
          component: () => import('@/views/Plugins.vue'),
          meta: { title: '插件系统' }
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
          path: 'workflow',
          name: 'WorkflowList',
          component: () => import('@/views/workflow/WorkflowList.vue'),
          meta: { title: '工作流' }
        },
        {
          path: 'workflow/:id',
          redirect: to => ({ path: `/workflow/run/${to.params.id as string}` })
        }
      ]
    },
    // ---- 教程中心（独立标签页打开，独立布局，不受主站左栏影响）----
    // 左栏两个吧 + 功能速查表；吧走 :board 参数复用同一套贴吧列表页。
    {
      path: '/tutorial',
      component: TutorialLayout,
      redirect: '/tutorial/board/experience',
      children: [
        {
          path: 'board/:board',
          name: 'TutorialBoard',
          component: () => import('@/views/ExperienceShares.vue'),
          meta: { title: '教程中心' }
        },
        {
          path: 'quickref',
          name: 'TutorialQuickRef',
          component: () => import('@/views/TutorialQuickRef.vue'),
          meta: { title: '功能速查表' }
        }
      ]
    },
    // 帖子详情：沉浸式贴吧楼层，仍走全屏布局（自带 tb-bar，塞进带左栏的布局要重做样式）
    {
      path: '/tutorial/post/:id',
      component: FullScreenLayout,
      children: [
        {
          path: '',
          name: 'TutorialPost',
          component: () => import('@/views/ExperienceShareDetail.vue'),
          meta: { title: '帖子详情' }
        }
      ]
    },
    // ---- 老路由兼容：经验分享已从主站左栏摘到教程中心，老地址一律重定向 ----
    { path: '/experience-shares', redirect: '/tutorial/board/experience' },
    {
      path: '/experience-shares/:id',
      redirect: to => ({ path: `/tutorial/post/${to.params.id as string}` })
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
      path: '/workflow/run/:id',
      component: FullScreenLayout,
      children: [
        {
          path: '',
          name: 'WorkflowRun',
          component: () => import('@/views/workflow/WorkflowRun.vue'),
          meta: { title: '工作流画布' }
        }
      ]
    }
  ]
})

export default router
