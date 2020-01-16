<template>
    <div class="wrapper">
        <v-head></v-head>
        <v-sidebar></v-sidebar>
        <div class="content-box" :class="{'content-collapse':isCollapse}">
            <v-tags></v-tags>
            <div class="content">
                <transition name="move" mode="out-in" appear>
                    <keep-alive :include="keepAlives">
                        <!-- 功能页面 -->
                        <router-view></router-view>
                    </keep-alive>
                </transition>
            </div>
        </div>
    </div>
</template>

<script>
    import vHead from './Header.vue'
    import vSidebar from './Sidebar.vue'
    import vTags from './Tags.vue'

    export default {
        components:{
            vHead, vSidebar, vTags
        },
        computed: {
            isCollapse() {
                return this.$store.getters.getScollapse
            },
            keepAlives(){
                return this.$store.getters.getKeepAlives
            }
        }
    }
</script>
