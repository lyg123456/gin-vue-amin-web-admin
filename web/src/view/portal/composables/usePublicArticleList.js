import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { getPublicArticleList } from '@/api/publicArticle'
import { getUrl } from '@/utils/image'

/**
 * 用户端公开文章列表（后台「内容 → 文章」中已发布），无需登录。
 */
export function usePublicArticleList(options = {}) {
  const { initialPageSize = 10 } = options
  const router = useRouter()
  const list = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(initialPageSize)
  const keyword = ref('')
  const loading = ref(true)

  const coverSrc = (url) => getUrl(url)

  const load = async () => {
    loading.value = true
    try {
      const res = await getPublicArticleList({
        page: page.value,
        pageSize: pageSize.value,
        keyword: keyword.value || undefined
      })
      if (res.code === 0) {
        list.value = res.data.list || []
        total.value = res.data.total || 0
        page.value = res.data.page || 1
        pageSize.value = res.data.pageSize || initialPageSize
      }
    } finally {
      loading.value = false
    }
  }

  const reload = () => {
    page.value = 1
    load()
  }

  const onPage = (p) => {
    page.value = p
    load()
  }

  const goArticle = (slug) => {
    router.push({ name: 'WebArticle', params: { slug } })
  }

  return {
    list,
    total,
    page,
    pageSize,
    keyword,
    loading,
    coverSrc,
    load,
    reload,
    onPage,
    goArticle
  }
}
