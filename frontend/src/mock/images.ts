import type { ImageAsset } from '../types/image'

const image = (
  id: number,
  name: string,
  url: string,
  size: number,
  createdAt: string,
  source: ImageAsset['source'] = 'upload',
): ImageAsset => ({
  id,
  name,
  url: `${url}?auto=format&fit=crop&w=900&h=520&q=82`,
  size,
  source,
  createdAt,
})

export const mockImages: ImageAsset[] = [
  image(1, 'docker-compose.png', 'https://images.unsplash.com/photo-1605745341112-85968b19335b', 128, '2026-09-24 14:32', 'article'),
  image(2, 'linux-tutorial.png', 'https://images.unsplash.com/photo-1629654297299-c8506221ca97', 256, '2026-09-20 10:15', 'article'),
  image(3, 'python-code.png', 'https://images.unsplash.com/photo-1526379095098-d400fd0bf935', 198, '2026-09-15 09:27', 'article'),
  image(4, 'nginx-config.png', 'https://images.unsplash.com/photo-1558494949-ef010cbdcc31', 142, '2026-09-12 11:03', 'upload'),
  image(5, 'mysql-database.png', 'https://images.unsplash.com/photo-1544383835-bda2bc66a55d', 176, '2026-09-10 14:22', 'article'),
  image(6, 'git-commands.png', 'https://images.unsplash.com/photo-1556075798-4825dfaaf498', 134, '2026-09-08 16:37', 'article'),
  image(7, 'server-room.jpg', 'https://images.unsplash.com/photo-1558494949-ef010cbdcc31', 320, '2026-09-05 10:19', 'upload'),
  image(8, 'code-editor.png', 'https://images.unsplash.com/photo-1515879218367-8466d910aaa4', 210, '2026-09-03 15:28', 'article'),
  image(9, 'cloud-architecture.png', 'https://images.unsplash.com/photo-1451187580459-43490279c0fa', 165, '2026-09-01 09:14', 'article'),
  image(10, 'work-desk.jpg', 'https://images.unsplash.com/photo-1496181133206-80ce9b88a853', 287, '2026-08-28 17:46', 'upload'),
  image(11, 'keyboard.jpg', 'https://images.unsplash.com/photo-1511467687858-23d96c32e4ae', 156, '2026-08-25 13:22', 'system'),
  image(12, 'landscape.jpg', 'https://images.unsplash.com/photo-1500534623283-312aade485b7', 412, '2026-08-22 16:30', 'upload'),
  image(13, 'infrastructure.png', 'https://images.unsplash.com/photo-1558494949-ef010cbdcc31', 198, '2026-08-20 11:07', 'article'),
  image(14, 'python-example.png', 'https://images.unsplash.com/photo-1516116216624-53e697fedbea', 143, '2026-08-18 09:53', 'article'),
  image(15, 'workspace.png', 'https://images.unsplash.com/photo-1497366754035-f200968a6e72', 102, '2026-08-15 14:18', 'system'),
]
