import { InjectionKey } from '@nuxtjs/composition-api'
import { PodStore } from '../../store/pod'

const PodStoreKey: InjectionKey<PodStore> = Symbol('PodStore')
export default PodStoreKey
