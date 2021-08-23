import { InjectionKey } from '@nuxtjs/composition-api'
import { SnackBarStore } from '../../store/snackbar'

const SnackBarStoreKey: InjectionKey<SnackBarStore> = Symbol('snackBarStore')
export default SnackBarStoreKey
