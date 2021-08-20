import { InjectionKey } from '@nuxtjs/composition-api'
import { StorageClassStore } from '../../store/storageclass'

const StorageClassStoreKey: InjectionKey<StorageClassStore> =
  Symbol('StorageClassStore')
export default StorageClassStoreKey
