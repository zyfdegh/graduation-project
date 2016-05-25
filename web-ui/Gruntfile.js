'use strict';

module.exports = function(grunt) {
	// load all grunt tasks
	require('matchdep').filterDev('grunt-*').forEach(grunt.loadNpmTasks);
	// configurable paths
	var moduleConfig = {
		src: 'portal-ui',
		test: 'test',
		dist: 'target/portal-ui',
		target: 'target',
		server: 'portal-server',
		deps: 'node_modules',
		artifact: 'linker-portal.zip'
	};

	grunt.initConfig({
		module: moduleConfig,
		pkg: grunt.file.readJSON('package.json'),
		clean: {
			dist: {
				files: [{
					dot: true,
					src: [
						'<%= module.target %>',
						'<%= module.artifact %>',
						'! <%= module.target %> /.git'
					]
				}]
			}
		},
		useminPrepare: {
			html: ['<%= module.src %>/index.html', '<%= module.src %>/login.html'],
			options: {
				dest: '<%= module.dist %>'
			}

		},
		usemin: {
			html: ['<%= module.dist %>/{,*/}*.html'],
			css: ['<%= module.dist %>/dist/{,*/}*.css'],
			options: {
				dirs: ['<%= module.dist %>']
			}
		},
		htmlmin: {
			dist: {
				options: {

				},
				files: [{
					expand: true,
					cwd: '<%= module.src %>',
					src: ['templates/**/*.html'],
					dest: '<%= module.dist %>'
				}, {
					expand: true,
					cwd: '<%= module.src %>',
					src: ['*.html'],
					dest: '<%= module.dist %>'
				}]
			}
		},
		cssmin: {
			css: {
				src: '<%= module.src %>/css/themes/**/*.css',
				dest: '<%= module.dist %>/css/themes/**/*.css'
			}
		},
		concat: {

		},
		// Put files not handled in other tasks here
		copy: {
			dist: {
				files: [{
					expand: true,
					dot: true,
					cwd: '<%= module.src %>',
					dest: '<%= module.dist %>',
					src: [
						'*.{ico,png,txt}',
						'locales/**/*',
						'conf/**/*',
						'css/**/*',
						'js/libs/**/*',
						'js/non-angular/i18n/**/*'
						// '!css/themes/**/*'
					]
				}, {
					expand: true,
					dot: true,
					cwd: '',
					dest: '<%= module.target %>',
					src: [
						'<%= module.server %>/**/*',
						'package.json'
					]
				}, {
					expand: true,
					dot: true,
					cwd: '',
					dest: '<%= module.target %>',
					src: [
						'node_modules/body-parser/**',
						'node_modules/connect-multiparty/**',
						'node_modules/connect-redis/**',
						'node_modules/cookie-parser/**',
						'node_modules/express/**',
						'node_modules/express-session/**',
						'node_modules/konphyg/**',
						'node_modules/node-zookeeper-client/**',
						'node_modules/redis-sentinel-client/**',
						'node_modules/request/**',
						'node_modules/sugar/**',
						'node_modules/winston/**'
					]
				}]
			}
		},
		ngmin: {

		},
		ngtemplates: {

		},
		uglify: {
			dist: {
				files: {
					'<%= module.dist %>/js/scripts.min.js': [
						'<%= module.dist %>/js/scripts.min.js'
					],
					// FIXME: for angular error
					'<%= module.dist %>/js/login.min.js': [
						'<%= module.dist %>/js/login.js'
					]
				}
			}
		},
		rev: {
			dist: {
				files: {
					src: [
						// '<%= module.dist %>/js/{,*/}*.js'
						'<%= module.dist %>/js/*.js'
					]
				}
			}
		},
		compress: {
			main: {
				options: {
					archive: '<%= module.artifact %>'
				},
				files: [ // path下的所有目录和文件
					{
						cwd: '<%= module.target %>',
						expand: true,
						src: [
							'**'
						],
						dest: ''
					}
				]
			}
		},
		manifest: {

		}
	});

	grunt.registerTask('build', [
		'useminPrepare',
		'copy',
		'htmlmin',
		// 'cssmin',
		'concat',
		'uglify:dist',
		'rev',
		'usemin',
		'compress'
	]);

	grunt.registerTask('default', [
		'clean',
		'build'
	]);

}
