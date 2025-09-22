package com.elgris.usersapi.services;

import com.elgris.usersapi.models.User;
import com.elgris.usersapi.repository.UserRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.stereotype.Service;

import java.util.Optional;

@Service
public class UserCacheService {

    private static final String USER_CACHE_PREFIX = "user:";

    @Autowired
    private UserRepository userRepository;

    @Autowired
    private RedisTemplate<String, User> redisTemplate;

    public static class UserCacheResponse {
        public User user;
        public String source;
        public long elapsedMillis;

        public UserCacheResponse(User user, String source, long elapsedMillis) {
            this.user = user;
            this.source = source;
            this.elapsedMillis = elapsedMillis;
        }
    }

    public Optional<UserCacheResponse> getUserByUsername(String username) {
        String key = USER_CACHE_PREFIX + username;
        long start = System.currentTimeMillis();

        User cachedUser = redisTemplate.opsForValue().get(key);
        if (cachedUser != null) {
            long elapsed = System.currentTimeMillis() - start;
            return Optional.of(new UserCacheResponse(cachedUser, "redis", elapsed));
        }

        User user = userRepository.findByUsername(username);
        if (user != null) {
            redisTemplate.opsForValue().set(key, user);
        }
        long elapsed = System.currentTimeMillis() - start;
        return Optional.of(new UserCacheResponse(user, "database", elapsed));
    }

    public User saveUser(User user) {
        User saved = userRepository.save(user);
        String key = USER_CACHE_PREFIX + saved.getUsername();
        redisTemplate.opsForValue().set(key, saved);
        return saved;
    }

    public void deleteUser(String username) {
        userRepository.deleteByUsername(username);
        String key = USER_CACHE_PREFIX + username;
        redisTemplate.delete(key);
    }
}